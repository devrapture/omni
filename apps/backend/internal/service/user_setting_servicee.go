package service

import (
	"context"
	"errors"

	apperrors "github.com/devrapture/omni/internal/errors"
	"github.com/devrapture/omni/internal/models"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSettingService interface {
	GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSetting, error)
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, userSetting *models.UserSetting) error
	DeleteUserKey(ctx context.Context, userID uuid.UUID) error
}

type userSettingService struct {
	repo repositories.UserSettingRepository
}

func NewUserSettingService(repo repositories.UserSettingRepository) UserSettingService {
	return &userSettingService{
		repo: repo,
	}
}

func (s *userSettingService) GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSetting, error) {
	settings, err := s.repo.FindByUserID(ctx, userID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrSettingsNotFound
		}
		return nil, err
	}
	return settings, nil
}

func (s *userSettingService) UpdateUserSettings(ctx context.Context, userID uuid.UUID, userSetting *models.UserSetting) error {
	return s.repo.Upsert(ctx, userID, userSetting)
}

func (s *userSettingService) DeleteUserKey(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteAPIKey(ctx, userID)
}
