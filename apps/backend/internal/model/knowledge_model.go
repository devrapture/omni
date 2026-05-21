package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type SourceType string

const (
	SourceTypeText SourceType = "text" // manually typed text
	SourceTypePDF  SourceType = "pdf"
	SourceTypeDocx SourceType = "docx"
	SourceTypeCSV  SourceType = "csv"
	SourceTypeXLSX SourceType = "xlsx"
	SourceTypeWeb  SourceType = "web"
)

const (
	DefaultEmbeddingModel     = "gemini-embedding-001"
	DefaultEmbeddingDimension = 768
)

type BusinessKnowledge struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Title      string     `json:"title" gorm:"type:text;not null"`
	Content    string     `json:"content" gorm:"type:text;not null"`
	SourceName string     `json:"source_name" gorm:"type:text;not null"` // filename or website url
	SourceType SourceType `json:"source_type" gorm:"type:text;not null"` // where this chunk came from? (text,pdf,docx,csv,xlsx,web)
	ChunkIndex int        `json:"chunk_index" gorm:"default:0"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`

	Embedding      pgvector.Vector `json:"-" gorm:"type:vector(768)"`
	EmbeddingModel string          `json:"embedding_model" gorm:"type:text;not null"`

	BusinessID uuid.UUID `json:"business_id" gorm:"type:uuid;not null"`
	Business   Business  `json:"-" gorm:"foreignKey:BusinessID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

func (b *BusinessKnowledge) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if b.EmbeddingModel == "" {
		b.EmbeddingModel = DefaultEmbeddingModel
	}
	return nil
}
