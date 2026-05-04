package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type BusinessKnowledge struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Title      string    `json:"title" gorm:"type:text;not null"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	SourceName string    `json:"source_name" gorm:"type:text;not null"` // filename or website url
	SourceType string    `json:"source_type" gorm:"type:text;not null"` // where this chunk came from? (text,pdf,docx,csv,xlsx,web)
	ChunkIndex int       `json:"chunk_index" gorm:"default:0"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	Embedding  pgvector.Vector
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User       User           `json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-"`
}

func (b *BusinessKnowledge) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
