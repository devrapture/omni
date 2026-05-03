package main

import (
	"fmt"
	"log"

	"github.com/devrapture/omni/internal/config"
	"github.com/devrapture/omni/internal/database"
	"github.com/devrapture/omni/routes"
	"github.com/devrapture/omni/utils"
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

	addr := fmt.Sprintf(":%s", cfg.Port)

	log.Printf("Server starting on %s", addr)
	r := routes.Setup(db)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server %v", err)
	}
}
