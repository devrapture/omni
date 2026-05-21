package service

import (
	"context"
	"fmt"

	"github.com/devrapture/omni/internal/model"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type EmbeddingService interface {
	EmbedDocument(ctx context.Context, text string) ([]float32, error)
	EmbedQuestion(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

type BusinessService interface {
	CreateBusiness(ctx context.Context, businessName string, userID uuid.UUID) (*model.Business, error)
	GetBusiness(ctx context.Context, id uuid.UUID) (*model.Business, error)
	GetKnowledge(ctx context.Context, businessID uuid.UUID) ([]model.BusinessKnowledge, error)
	DeleteBySource(ctx context.Context, businessID uuid.UUID, sourceName string) error
	IngestText(ctx context.Context, businessID uuid.UUID, title, content, sourceName string, sourceType model.SourceType) (int, error)
}

type businessService struct {
	businessRepository  repositories.BusinessRepository
	knowledgeRepository repositories.KnowledgeRepository
	embeddingService    EmbeddingService
	logger              *zap.Logger
}

func NewBusinessService(br repositories.BusinessRepository, kr repositories.KnowledgeRepository, es EmbeddingService, logger *zap.Logger) BusinessService {
	return &businessService{
		businessRepository:  br,
		knowledgeRepository: kr,
		embeddingService:    es,
		logger:              logger,
	}
}

func (s *businessService) CreateBusiness(ctx context.Context, businessName string, userID uuid.UUID) (*model.Business, error) {
	business, err := s.businessRepository.CreateBusiness(ctx, businessName, userID)
	if err != nil {
		return nil, err
	}

	return business, nil
}

func (s *businessService) GetBusiness(ctx context.Context, id uuid.UUID) (*model.Business, error) {
	return s.businessRepository.FindByID(ctx, id)
}

func (s *businessService) GetKnowledge(ctx context.Context, businessID uuid.UUID) ([]model.BusinessKnowledge, error) {
	return s.knowledgeRepository.FindByBusinessID(ctx, businessID)
}

func (s *businessService) DeleteBySource(ctx context.Context, businessID uuid.UUID, sourceName string) error {
	return s.knowledgeRepository.DeleteBySource(ctx, businessID, sourceName)
}

func (s *businessService) IngestText(ctx context.Context, businessID uuid.UUID, title, content, sourceName string, sourceType model.SourceType) (int, error) {
	s.logger.Info("Starting text ingestion", zap.String("businessID", businessID.String()), zap.String("sourceName", sourceName), zap.String("sourceType", string(sourceType)))
	cfg := DefaultChunkConfig()
	chunks := ChunkText(content, cfg)

	if len(chunks) == 0 {
		return 0, fmt.Errorf("no content could be extracted from the provided text")
	}

	s.logger.Info("Text chunked", zap.Int("num_chunks", len(chunks)))
	embeddings, err := s.batchEmbed(ctx, chunks)
	if err != nil {
		return 0, fmt.Errorf("embedding failed: %w", err)
	}

	records := make([]model.BusinessKnowledge, len(chunks))
	for i, chunk := range chunks {
		records[i] = model.BusinessKnowledge{
			BusinessID:     businessID,
			Title:          title,
			Content:        chunk,
			SourceType:     sourceType,
			SourceName:     sourceName,
			ChunkIndex:     i,
			IsActive:       true,
			Embedding:      pgvector.NewVector(embeddings[i]),
			EmbeddingModel: model.DefaultEmbeddingModel,
		}
	}

	if err := s.knowledgeRepository.CreateChunks(ctx, records); err != nil {
		return 0, fmt.Errorf("failed to store knowledge chunks: %w", err)
	}
	return len(records), nil
}

func (s *businessService) batchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	const batchSize = 50
	var allEmbeddings [][]float32
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := s.embeddingService.EmbedBatch(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("batch %d-%d embeddings failed: %w", i, end, err)
		}
		allEmbeddings = append(allEmbeddings, embeddings...)
	}
	return allEmbeddings, nil
}
