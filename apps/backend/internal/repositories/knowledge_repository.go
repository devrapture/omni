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

	// CreateChunks stores multiple chunks in a single transaction.
	// Use this when ingesting a file or crawled page.
	CreateChunks(ctx context.Context, chunks []model.BusinessKnowledge) error

	// FindByBusinessID returns all chunks for a user (for listing).
	FindByBusinessID(ctx context.Context, businessID uuid.UUID) ([]model.BusinessKnowledge, error)

	// DeleteBySource removes all chunks from a specific source (e.g., when re-uploading a file).
	DeleteBySource(ctx context.Context, businessID uuid.UUID, sourceName string) error 
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

// CreateChunks uses a transaction to insert all chunks atomically.
// If any chunk fails, the entire batch is rolled back.
// This prevents partial uploads (e.g., half a PDF being indexed).
func (r *knowledgeRepository) CreateChunks(ctx context.Context, chunks []model.BusinessKnowledge) error {
	if len(chunks) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		batchSize := 100
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
	})
}

func (r *knowledgeRepository) FindByBusinessID(ctx context.Context, businessID uuid.UUID) ([]model.BusinessKnowledge, error) {
	var entries []model.BusinessKnowledge
	err := r.db.WithContext(ctx).Where("business_id = ? AND is_active = true", businessID).Order("source_name ASC, chunk_index ASC").Find(&entries).Error
	return entries, err
}

func (r *knowledgeRepository) DeleteBySource(ctx context.Context, businessID uuid.UUID, sourceName string) error {
	return r.db.WithContext(ctx).Where("business_id = ? AND source_name = ?", businessID, sourceName).Delete(&model.BusinessKnowledge{}).Error
}
