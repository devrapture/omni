package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadJobStatus string

const (
	UploadJobQueued     UploadJobStatus = "queued"
	UploadJobProcessing UploadJobStatus = "processing"
	UploadJobCompleted  UploadJobStatus = "completed"
	UploadJobFailed     UploadJobStatus = "failed"
)

type UploadJob struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status     UploadJobStatus `json:"status" gorm:"type:text;not null;default:queued"`
	UserID     uuid.UUID       `json:"user_id" gorm:"type:uuid;not null;index"`
	ObjectKey  string          `json:"object_key" gorm:"type:text;not null"`
	SourceName string          `json:"source_name" gorm:"type:text;not null"`
	SourceType string          `json:"source_type" gorm:"type:text;not null"`
	Content    string          `json:"content,omitempty" gorm:"type:text"`
	Error      string          `json:"error,omitempty" gorm:"type:text"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func (u *UploadJob) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
