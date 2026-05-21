package main

import (
	"context"
	"fmt"
	"log"

	"github.com/devrapture/omni/internal/config"
	"github.com/devrapture/omni/internal/database"
	handlers "github.com/devrapture/omni/internal/handler"
	"github.com/devrapture/omni/internal/integrations/gemini"
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

	// Services
	userSvc := service.NewUserService(cfg, userRepo)
	userSettingsSvc := service.NewUserSettingService(userSettingRepo, cfg)
	parserSvc := service.NewParserService()
	embeddingSvc, err := gemini.NewEmbeddingClient(context.Background(), cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini embedding client: %v", err)
	}
	businessSvc := service.NewBusinessService(businessRepo, knowledgeRepo, embeddingSvc, logger)

	asynqClient := asynq.NewClient(queue.RedisClientOpt(cfg))
	defer asynqClient.Close()

	// Handlers
	authHandler := handlers.NewAuthHandler(userSvc)
	userSettingHandler := handlers.NewUserSettingsHandler(userSettingsSvc)
	fileUploadHandler := handlers.NewFileUploadHandler(cfg, parserSvc, asynqClient, uploadJobRepo, r2Storage, logger)
	businessHandler := handlers.NewBusinessHandler(businessSvc)

	deps := routes.HandlerDependencies{
		AuthHandler:         authHandler,
		UserSettingsHandler: userSettingHandler,
		FileUploadHandler:   fileUploadHandler,
		BusinessHandler:     businessHandler,
	}

	addr := fmt.Sprintf(":%s", cfg.Port)

	r := routes.Setup(db, deps, cfg, logger)
	logger.Info("Server starting", zap.String("addr", addr))

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server %v", err)
	}
}
