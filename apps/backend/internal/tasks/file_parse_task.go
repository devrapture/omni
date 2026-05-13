package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devrapture/omni/internal/model"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/devrapture/omni/internal/service"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const TypeFileParse = "file:parse"

type FileParsePayload struct {
	JobID    uuid.UUID `json:"job_id"`
	UserID   uuid.UUID `json:"user_id"`
	FilePath string    `json:"file_path"`
}

func NewFileParseTask(jobID, userID uuid.UUID, filePath string) (*asynq.Task, error) {
	payload, err := json.Marshal(FileParsePayload{
		JobID:    jobID,
		UserID:   userID,
		FilePath: filePath,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeFileParse, payload), nil
}

func HandleFileParseTask(uploadJobRepo repositories.UploadJobRepository, parser *service.ParserService, logger *zap.Logger) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload FileParsePayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		if err := uploadJobRepo.UpdateJob(ctx, payload.JobID, model.UploadJobProcessing, ""); err != nil {
			return err
		}

		content, sourceType, err := parser.Parse(payload.FilePath)
		if err != nil {
			_ = uploadJobRepo.UpdateJob(ctx, payload.JobID, model.UploadJobFailed, err.Error())
			return err
		}

		uploadJobRepo.MarkCompleted(ctx, payload.JobID, content, sourceType)

		logger.Info("file parsed", zap.Any("user_id", payload.UserID), zap.String("file_path", payload.FilePath), zap.String("source_type", sourceType), zap.Int("content length", len(content)))

		// Store content here:
		// - save to business_knowledges
		// - chunk content
		// - generate embeddings
		// - mark upload complete
		return nil
	}
}

func FileParseOptions() []asynq.Option {
	return []asynq.Option{
		asynq.Queue("file_parsing"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}
}
