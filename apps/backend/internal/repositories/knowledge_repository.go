package repositories

import (
	"context"

	"github.com/devrapture/omni/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KnowledgeRepository interface {
	// CreateChunk stores a single chunk with its embedding.
	CreateChunk(ctx context.Context, businessKnowlege *model.BusinessKnowledge) error

	// ReplaceChunksBySource replaces all chunks for the same business/source atomically.
	// Use this for idempotent file ingestion and task retries.
	ReplaceChunksBySource(ctx context.Context, businessID uuid.UUID, sourceName string, chunks []model.BusinessKnowledge) error

	// FindByBusinessID returns all chunks for a user (for listing).
	FindByBusinessID(ctx context.Context, businessID uuid.UUID) ([]model.BusinessKnowledge, error)

	// DeleteBySource removes all chunks from a specific source (e.g., when re-uploading a file).
	DeleteUserSource(ctx context.Context, businessID uuid.UUID, sourceName string) error
}

type knowledgeRepository struct {
	db *gorm.DB
}

func NewKnowledgeRepository(db *gorm.DB) KnowledgeRepository {
	return &knowledgeRepository{db: db}
}

func (r *knowledgeRepository) CreateChunk(ctx context.Context, businessKnowlege *model.BusinessKnowledge) error {
	return r.db.WithContext(ctx).Create(businessKnowlege).Error
}

func (r *knowledgeRepository) ReplaceChunksBySource(ctx context.Context, businessID uuid.UUID, sourceName string, chunks []model.BusinessKnowledge) error {
	if len(chunks) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().
			Where("business_id = ? AND source_name = ?", businessID, sourceName).
			Delete(&model.BusinessKnowledge{}).Error; err != nil {
			return err
		}
		return createChunksInBatches(tx, chunks)
	})
}

func (r *knowledgeRepository) FindByBusinessID(ctx context.Context, businessID uuid.UUID) ([]model.BusinessKnowledge, error) {
	var entries []model.BusinessKnowledge
	err := r.db.WithContext(ctx).Where("business_id = ? AND is_active = true", businessID).Order("source_name ASC, chunk_index ASC").Find(&entries).Error
	return entries, err
}

func (r *knowledgeRepository) DeleteUserSource(ctx context.Context, businessID uuid.UUID, sourceName string) error {
	return r.db.WithContext(ctx).Where("business_id = ? AND source_name = ?", businessID, sourceName).Delete(&model.BusinessKnowledge{}).Error
}

func createChunksInBatches(tx *gorm.DB, chunks []model.BusinessKnowledge) error {
	const batchSize = 100
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		if err := tx.Create(chunks[i:end]).Error; err != nil {
			return err
		}
	}
	return nil
}
