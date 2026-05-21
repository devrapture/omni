package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/devrapture/omni/internal/model"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/storage"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const TypeFileParse = "file:parse"

type FileParsePayload struct {
	JobID      uuid.UUID `json:"job_id"`
	UserID     uuid.UUID `json:"user_id"`
	BusinessID uuid.UUID `json:"business_id"`
	ObjectKey  string    `json:"object_key"`
	SourceName string    `json:"source_name"`
	Title      string    `json:"title"`
}

func NewFileParseTask(jobID, userID uuid.UUID, objectKey string) (*asynq.Task, error) {
	payload, err := json.Marshal(FileParsePayload{
		JobID:     jobID,
		UserID:    userID,
		ObjectKey: objectKey,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeFileParse, payload), nil
}

func HandleFileParseTask(uploadJobRepo repositories.UploadJobRepository, parser *service.ParserService, businessSvc service.BusinessService, r2 *storage.R2Storage, logger *zap.Logger) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload FileParsePayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		if err := uploadJobRepo.UpdateJob(ctx, payload.JobID, model.UploadJobProcessing, ""); err != nil {
			return handleParseFailure(ctx, r2, payload.ObjectKey, err, payload.JobID, uploadJobRepo)
		}

		localPath, cleanup, err := downloadR2ObjectToTempFile(ctx, r2, payload.ObjectKey)
		if err != nil {
			return handleParseFailure(ctx, r2, payload.ObjectKey, err, payload.JobID, uploadJobRepo)
		}

		defer cleanup()

		content, sourceType, err := parser.Parse(localPath)
		if err != nil {
			return handleParseFailure(ctx, r2, payload.ObjectKey, err, payload.JobID, uploadJobRepo)
		}

		chunks, err := businessSvc.IngestText(ctx, payload.BusinessID, payload.Title, content, payload.SourceName, sourceType)

		if err != nil {
			return handleParseFailure(ctx, r2, payload.ObjectKey, err, payload.JobID, uploadJobRepo)
		}

		logger.Info("file ingested into business knowledge",
			zap.String("business_id", payload.BusinessID.String()),
			zap.String("source_name", payload.SourceName),
			zap.Any("source_type", sourceType),
			zap.Int("chunks", chunks),
		)
		
		if err := uploadJobRepo.MarkCompleted(ctx, payload.JobID, content, sourceType); err != nil {
			return err
		}

		if err := r2.DeleteObject(ctx, payload.ObjectKey); err != nil {
			logger.Warn("failed to delete parsed file from r2", zap.String("object_key", payload.ObjectKey), zap.Error(err))
		}

		logger.Info("file parsed and deleted from r2", zap.Any("user_id", payload.UserID), zap.String("file_path", payload.ObjectKey), zap.Any("source_type", sourceType), zap.Int("content length", len(content)))

		// Store content here:
		// - save to business_knowledges
		// - chunk content
		// - generate embeddings
		// - mark upload complete
		return nil
	}
}

func downloadR2ObjectToTempFile(ctx context.Context, r2 *storage.R2Storage, objectKey string) (string, func(), error) {
	body, err := r2.GetObject(ctx, objectKey)
	if err != nil {
		return "", nil, err
	}
	defer body.Close()

	ext := strings.ToLower(filepath.Ext(objectKey))

	tmpFile, err := os.CreateTemp("", "upload-*"+ext)
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		_ = os.Remove(tmpFile.Name())
	}

	if _, err := io.Copy(tmpFile, body); err != nil {
		_ = tmpFile.Close()
		cleanup()
		return "", nil, err
	}

	if err := tmpFile.Close(); err != nil {
		cleanup()
		return "", nil, err
	}

	return tmpFile.Name(), cleanup, nil
}

func handleParseFailure(ctx context.Context, r2 *storage.R2Storage, objectKey string, parseErr error, jobID uuid.UUID, uploadJobRepo repositories.UploadJobRepository) error {
	retried, hasRetryCount := asynq.GetRetryCount(ctx)
	maxRetry, hasMaxRetry := asynq.GetMaxRetry(ctx)

	if hasRetryCount && hasMaxRetry && retried >= maxRetry {
		if updateErr := uploadJobRepo.UpdateJob(ctx, jobID, model.UploadJobFailed, parseErr.Error()); updateErr != nil {
			return fmt.Errorf("mark upload job failed: %w", updateErr)
		}

		if deleteErr := r2.DeleteObject(ctx, objectKey); deleteErr != nil {
			return fmt.Errorf("parse failed: %v; also failed to delete r2 object: %w", parseErr, deleteErr)
		}
	}

	return parseErr
}

func FileParseOptions() []asynq.Option {
	return []asynq.Option{
		asynq.Queue("file_parsing"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}
}
