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

func (r *uploadJobRepository) UpdateJob(ctx context.Context, id uuid.UUID, status model.UploadJobStatus, errorMessage string) error {
	return r.db.WithContext(ctx).Model(&model.UploadJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": status,
		"error":  errorMessage,
	}).Error
}

func (r *uploadJobRepository) MarkCompleted(ctx context.Context, id uuid.UUID, content, sourceType, errorMessage string) error {
	return r.db.WithContext(ctx).Model(&model.UploadJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      model.UploadJobCompleted,
		"content":     content,
		"source_type": sourceType,
		"error":       "",
	}).Error
}
