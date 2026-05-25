package service

import (
	"context"

	"github.com/devrapture/omni/internal/config"
	apperrors "github.com/devrapture/omni/internal/errors"
	"github.com/devrapture/omni/internal/integrations/gemini"
	"github.com/devrapture/omni/internal/model"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/devrapture/omni/internal/utils"
	"github.com/google/uuid"
)

type EmbeddingProvider interface {
	ForUser(ctx context.Context, userID uuid.UUID) (EmbeddingService, error)
}

type embeddingProvider struct {
	settingsRepo repositories.UserSettingRepository
	cfg          *config.Config
}

func NewEmbeddingProvider(settingsRepo repositories.UserSettingRepository, cfg *config.Config) EmbeddingProvider {
	return &embeddingProvider{
		settingsRepo: settingsRepo,
		cfg:          cfg,
	}
}

func (p *embeddingProvider) ForUser(ctx context.Context, userID uuid.UUID) (EmbeddingService, error) {
	settings, err := p.settingsRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if settings.Mode == model.AIKeyModePlatform || settings.APIKeyEncrypted == "" {
		return gemini.NewEmbeddingClient(ctx, p.cfg.GeminiAPIKey)
	}

	if settings.Provider != model.AIProviderGemini {
		return nil, apperrors.ErrInvalidGeminiKey
	}

	apiKey, err := utils.DecryptText(settings.APIKeyEncrypted, p.cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}
	return gemini.NewEmbeddingClient(ctx, apiKey)
}
