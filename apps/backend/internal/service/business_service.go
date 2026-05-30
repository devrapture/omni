package service

import (
	"context"
	"fmt"
	"time"

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
	GetKnowledgeForUser(ctx context.Context, businessID, userID uuid.UUID, sourceTypes []model.SourceType) ([]model.BusinessKnowledge, error)
	ListSources(ctx context.Context, businessID, userID uuid.UUID, sourceTypeQuery string) ([]KnowledgeSourceSummary, error)
	AddText(ctx context.Context, businessID, userID uuid.UUID, title, content string) (int, error)
	DeleteBySource(ctx context.Context, businessID, userID uuid.UUID, sourceName string) error
	IngestText(ctx context.Context, businessID, userID uuid.UUID, title, content, sourceName string, sourceType model.SourceType) (int, error)
}

type KnowledgeSourceSummary struct {
	SourceName string           `json:"source_name"`
	SourceType model.SourceType `json:"source_type"`
	ChunkCount int              `json:"chunk_count"`
	CreatedAt  time.Time        `json:"created_at"`
}

type businessService struct {
	businessRepository  repositories.BusinessRepository
	knowledgeRepository repositories.KnowledgeRepository
	embeddingProvider   EmbeddingProvider
	logger              *zap.Logger
}

func NewBusinessService(br repositories.BusinessRepository, kr repositories.KnowledgeRepository, ep EmbeddingProvider, logger *zap.Logger) BusinessService {
	return &businessService{
		businessRepository:  br,
		knowledgeRepository: kr,
		embeddingProvider:   ep,
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

func (s *businessService) GetKnowledgeForUser(ctx context.Context, businessID, userID uuid.UUID, sourceTypes []model.SourceType) ([]model.BusinessKnowledge, error) {
	if _, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID); err != nil {
		return nil, err
	}
	return s.knowledgeRepository.FindByBusinessID(ctx, businessID, sourceTypes)
}

func (s *businessService) ListSources(ctx context.Context, businessID, userID uuid.UUID, sourceTypeQuery string) ([]KnowledgeSourceSummary, error) {
	var sourceTypes []model.SourceType
	switch sourceTypeQuery {
	case "text":
		sourceTypes = []model.SourceType{model.SourceTypeText}
	case "file":
		sourceTypes = []model.SourceType{
			model.SourceTypePDF,
			model.SourceTypeDocx,
			model.SourceTypeCSV,
			model.SourceTypeXLSX,
		}
	case "web":
		sourceTypes = []model.SourceType{model.SourceTypeWeb}
	default: // "all" or anything unrecognised
		sourceTypes = nil
	}
	entries, err := s.GetKnowledgeForUser(ctx, businessID, userID, sourceTypes)
	if err != nil {
		return nil, err
	}

	sourceMap := make(map[string]*KnowledgeSourceSummary)
	for _, e := range entries {

		key := e.SourceName
		if _, exists := sourceMap[key]; !exists {
			sourceMap[key] = &KnowledgeSourceSummary{
				SourceName: e.SourceName,
				SourceType: e.SourceType,
				CreatedAt:  e.CreatedAt,
			}
		}
		sourceMap[key].ChunkCount++
	}

	sources := make([]KnowledgeSourceSummary, 0, len(sourceMap))
	for _, source := range sourceMap {
		sources = append(sources, *source)
	}
	return sources, nil
}

func (s *businessService) DeleteBySource(ctx context.Context, businessID, userID uuid.UUID, sourceName string) error {
	if _, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID); err != nil {
		return err
	}
	return s.knowledgeRepository.DeleteUserSource(ctx, businessID, sourceName)
}

func (s *businessService) AddText(ctx context.Context, businessID, userID uuid.UUID, title, content string) (int, error) {
	_, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID)
	if err != nil {
		return 0, err
	}
	const maxContentBytes = 10_000 // tune to your quota/latency budget
	if len(content) > maxContentBytes {
		return 0, fmt.Errorf("content too large: max %d bytes", maxContentBytes)
	}
	s.logger.Info("Starting text ingestion", zap.String("businessID", businessID.String()), zap.String("sourceName", title), zap.String("sourceType", string(model.SourceTypeText)))
	cfg := DefaultChunkConfig()
	chunks := ChunkText(content, cfg)

	if len(chunks) == 0 {
		return 0, fmt.Errorf("no content could be extracted from the provided text")
	}

	s.logger.Info("Text chunked", zap.Int("num_chunks", len(chunks)))
	embeddings, err := s.batchEmbed(ctx, chunks, userID)

	if err != nil {
		s.logger.Error("embedding failed", zap.Error(err))
		return 0, err
	}

	if len(embeddings) != len(chunks) {
		return 0, fmt.Errorf("embedding count mismatch: got %d embeddings for %d chunks", len(embeddings), len(chunks))
	}
	records := make([]model.BusinessKnowledge, len(chunks))
	for i, chunk := range chunks {
		records[i] = model.BusinessKnowledge{
			BusinessID:     businessID,
			Content:        chunk,
			SourceType:     model.SourceTypeText,
			SourceName:     title,
			ChunkIndex:     i,
			IsActive:       true,
			Embedding:      pgvector.NewVector(embeddings[i]),
			EmbeddingModel: model.DefaultEmbeddingModel,
		}
	}

	if err := s.knowledgeRepository.ReplaceChunksBySource(ctx, businessID, title, records); err != nil {
		return 0, fmt.Errorf("failed to store knowledge chunks: %w", err)
	}
	s.logger.Info("text successfully ingested into business knowledge",
		zap.String("business_id", businessID.String()),
		zap.String("source_name", title),
		zap.Any("source_type", model.SourceTypeText),
		zap.Int("chunks", len(records)),
	)
	return len(records), nil
}

func (s *businessService) IngestText(ctx context.Context, businessID, userID uuid.UUID, title, content, sourceName string, sourceType model.SourceType) (int, error) {
	s.logger.Info("Starting text ingestion", zap.String("businessID", businessID.String()), zap.String("sourceName", sourceName), zap.String("sourceType", string(sourceType)))
	cfg := DefaultChunkConfig()
	chunks := ChunkText(content, cfg)

	if len(chunks) == 0 {
		return 0, fmt.Errorf("no content could be extracted from the provided text")
	}

	s.logger.Info("Text chunked", zap.Int("num_chunks", len(chunks)), zap.Duration("latency_ms", time.Millisecond*120))
	embeddings, err := s.batchEmbed(ctx, chunks, userID)
	if err != nil {
		return 0, fmt.Errorf("embedding failed: %w", err)
	}

	if len(embeddings) != len(chunks) {
		return 0, fmt.Errorf("embedding count mismatch: got %d embeddings for %d chunks", len(embeddings), len(chunks))
	}

	records := make([]model.BusinessKnowledge, len(chunks))
	for i, chunk := range chunks {
		records[i] = model.BusinessKnowledge{
			BusinessID:     businessID,
			Content:        chunk,
			SourceType:     sourceType,
			SourceName:     sourceName,
			ChunkIndex:     i,
			IsActive:       true,
			Embedding:      pgvector.NewVector(embeddings[i]),
			EmbeddingModel: model.DefaultEmbeddingModel,
		}
	}

	if err := s.knowledgeRepository.ReplaceChunksBySource(ctx, businessID, sourceName, records); err != nil {
		return 0, fmt.Errorf("failed to store knowledge chunks: %w", err)
	}
	return len(records), nil
}

func (s *businessService) batchEmbed(ctx context.Context, texts []string, userID uuid.UUID) ([][]float32, error) {
	embeddingService, err := s.embeddingProvider.ForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	const batchSize = 50
	var allEmbeddings [][]float32
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := embeddingService.EmbedBatch(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("batch %d-%d embeddings failed: %w", i, end, err)
		}
		allEmbeddings = append(allEmbeddings, embeddings...)
	}
	return allEmbeddings, nil
}
