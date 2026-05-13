package main

import (
	"log"

	"github.com/devrapture/omni/internal/config"
	"github.com/devrapture/omni/internal/database"
	"github.com/devrapture/omni/internal/queue"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/tasks"
	"github.com/devrapture/omni/internal/utils"
	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration %v", err)
	}

	logger, err := utils.NewLogger(cfg.AppEnv == "development")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	db, err := database.ConnectDb(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	uploadJobRepo := repositories.NewUploadJobRepository(db)
	parserSvc := service.NewParserService()

	server := asynq.NewServer(
		queue.RedisClientOpt(cfg),
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"file_parsing": 10,
				"default":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeFileParse, tasks.HandleFileParseTask(uploadJobRepo, parserSvc, logger))

	logger.Info("Starting file parser worker")

	if err := server.Run(mux); err != nil {
		log.Fatalf("Failed to run worker: %v", err)
	}
}
