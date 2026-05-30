package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/devrapture/omni/internal/config"
	"github.com/devrapture/omni/internal/database"
	handlers "github.com/devrapture/omni/internal/handler"
	"github.com/devrapture/omni/internal/queue"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/devrapture/omni/internal/routes"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/storage"
	"github.com/devrapture/omni/internal/utils"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration %v", err)
	}

	isDevelopment := cfg.AppEnv == "development"

	logger, err := utils.NewLogger(isDevelopment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync() // flush buffered log entries on exit

	logger.Info("Starting Omnibot server", zap.String("env", cfg.AppEnv), zap.String("port", cfg.Port))

	db, err := database.ConnectDb(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r2Storage, err := storage.NewR2Storage(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to initialize R2 storage: %v", err)
	}

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	userSettingRepo := repositories.NewUserSettingRepository(db)
	uploadJobRepo := repositories.NewUploadJobRepository(db)
	businessRepo := repositories.NewBusinessRepository(db)
	knowledgeRepo := repositories.NewKnowledgeRepository(db)
	businessChannelSettingsRepo := repositories.NewBusinessChannelSettingsRepository(db)

	// Services
	userSvc := service.NewUserService(cfg, userRepo)
	userSettingsSvc := service.NewUserSettingService(userSettingRepo, cfg)
	parserSvc := service.NewParserService()
	businessChannelSettingService := service.NewBusinessChannelSettings(businessRepo, businessChannelSettingsRepo, logger, cfg)

	// ── Register Telegram webhooks for all active bots ─────────────────────
	// This runs synchronously before the HTTP server starts.
	// Why synchronous? We want to be sure all bots are registered before we
	// start accepting messages. If we did this async, there's a window where
	// a message could arrive before registration completes.
	//
	// GetWebhookInfo is called per-bot to skip already-registered webhooks,
	// so this is fast even with many businesses.
	logger.Info("Registering Telegram webhooks for all active bots...")

	webhookCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := businessChannelSettingService.RegisterAllTelegramWebhooks(webhookCtx, cfg.AppBaseUrl); err != nil {
		// Non-fatal: log the error but start the server anyway.
		// Individual bot failures are logged inside RegisterAllTelegramWebhooks.
		logger.Warn("Some Telegram webhook registrations failed", zap.Error(err))
	}

	embeddingProvider := service.NewEmbeddingProvider(userSettingRepo, cfg)
	businessSvc := service.NewBusinessService(businessRepo, knowledgeRepo, embeddingProvider, logger)

	asynqClient := asynq.NewClient(queue.RedisClientOpt(cfg))
	defer asynqClient.Close()

	// Handlers
	authHandler := handlers.NewAuthHandler(userSvc)
	userSettingHandler := handlers.NewUserSettingsHandler(userSettingsSvc)
	fileUploadHandler := handlers.NewFileUploadHandler(cfg, parserSvc, asynqClient, uploadJobRepo, businessRepo, r2Storage, logger)
	businessHandler := handlers.NewBusinessHandler(businessSvc)
	BusinessChannelSettingHandler := handlers.NewBusinessChannelSettingHandler(businessChannelSettingService)

	deps := routes.HandlerDependencies{
		AuthHandler:                   authHandler,
		UserSettingsHandler:           userSettingHandler,
		FileUploadHandler:             fileUploadHandler,
		BusinessHandler:               businessHandler,
		BusinessChannelSettingHandler: BusinessChannelSettingHandler,
	}

	addr := fmt.Sprintf(":%s", cfg.Port)

	r := routes.Setup(db, deps, cfg, logger)
	logger.Info("Server starting", zap.String("addr", addr))

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server %v", err)
	}
}
