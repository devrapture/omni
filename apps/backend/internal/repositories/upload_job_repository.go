package repositories

import (
	"context"

	"github.com/devrapture/omni/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadJobRepository interface {
	Create(ctx context.Context, job *model.UploadJob) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.UploadJob, error)
	FindByIDandUserID(ctx context.Context, userID, jobID uuid.UUID) (*model.UploadJob, error)
	UpdateJob(ctx context.Context, id uuid.UUID, status model.UploadJobStatus, errorMessage string) error
	ClaimQueuedJob(ctx context.Context, id uuid.UUID) error
	ReleaseProcessingJob(ctx context.Context, id uuid.UUID) error
	MarkCompleted(ctx context.Context, id uuid.UUID, content string, sourceType model.SourceType) error
}

type uploadJobRepository struct {
	db *gorm.DB
}

func NewUploadJobRepository(DB *gorm.DB) UploadJobRepository {
	return &uploadJobRepository{
		db: DB,
	}
}

func (r *uploadJobRepository) Create(ctx context.Context, job *model.UploadJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *uploadJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.UploadJob, error) {
	var job model.UploadJob
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *uploadJobRepository) FindByIDandUserID(ctx context.Context, userID, jobID uuid.UUID) (*model.UploadJob, error) {
	var job model.UploadJob
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", jobID, userID).First(&job).Error; err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *uploadJobRepository) UpdateJob(ctx context.Context, id uuid.UUID, status model.UploadJobStatus, errorMessage string) error {
	tx := r.db.WithContext(ctx).Model(&model.UploadJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": status,
		"error":  errorMessage,
	})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *uploadJobRepository) ClaimQueuedJob(ctx context.Context, id uuid.UUID) error {
	tx := r.db.WithContext(ctx).Model(&model.UploadJob{}).
		Where("id = ? AND status = ?", id, model.UploadJobQueued).
		Updates(map[string]interface{}{
			"status": model.UploadJobProcessing,
			"error":  "",
		})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *uploadJobRepository) ReleaseProcessingJob(ctx context.Context, id uuid.UUID) error {
	tx := r.db.WithContext(ctx).Model(&model.UploadJob{}).
		Where("id = ? AND status = ?", id, model.UploadJobProcessing).
		Updates(map[string]interface{}{
			"status": model.UploadJobQueued,
			"error":  "",
		})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *uploadJobRepository) MarkCompleted(ctx context.Context, id uuid.UUID, content string, sourceType model.SourceType) error {
	tx := r.db.WithContext(ctx).Model(&model.UploadJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      model.UploadJobCompleted,
		"content":     content,
		"source_type": sourceType,
		"error":       "",
	})
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
