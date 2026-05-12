package service

import (
	"context"
	"errors"
	"strings"

	"github.com/devrapture/omni/internal/config"
	"github.com/devrapture/omni/internal/dto"
	apperrors "github.com/devrapture/omni/internal/errors"
	"github.com/devrapture/omni/internal/model"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/devrapture/omni/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSettingService interface {
	GetUserSettings(ctx context.Context, userID uuid.UUID) (*model.UserSetting, error)
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, dto dto.UpdateUserSettingsDTO) error
	DeleteUserKey(ctx context.Context, userID uuid.UUID) error
}

type userSettingService struct {
	repo repositories.UserSettingRepository
	cfg  *config.Config
}

func NewUserSettingService(repo repositories.UserSettingRepository, cfg *config.Config) UserSettingService {
	if cfg == nil {
		panic("user setting service requires non-nil config")
	}
	return &userSettingService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *userSettingService) GetUserSettings(ctx context.Context, userID uuid.UUID) (*model.UserSetting, error) {
	settings, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrSettingsNotFound
		}
		return nil, err
	}
	return settings, nil
}

func (s *userSettingService) UpdateUserSettings(ctx context.Context, userID uuid.UUID, dto dto.UpdateUserSettingsDTO) error {
	setting := &model.UserSetting{
		UserID:   userID,
		Mode:     dto.Mode,
		Provider: dto.Provider,
	}
	if dto.Mode == model.AIKeyModeUserKey {
		apiKey := strings.TrimSpace(dto.APIKey)
		if apiKey == "" {
			return apperrors.ErrEmptyAPIKey
		}

		encryptedKey, err := utils.EncryptText(apiKey, s.cfg.EncryptionKey)

		if err != nil {
			return err
		}
		setting.APIKeyEncrypted = encryptedKey
	}
	return s.repo.Upsert(ctx, userID, setting)
}

func (s *userSettingService) DeleteUserKey(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteAPIKey(ctx, userID)
}
