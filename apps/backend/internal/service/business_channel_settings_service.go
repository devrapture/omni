package service

import (
	"context"

	"github.com/devrapture/omni/internal/model"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/google/uuid"
)

type BusinessChannelSetting interface {
	Get(ctx context.Context, businessID, userID uuid.UUID) (*model.BusinessChannelSetting, error)
}

type businessChannelSetting struct {
	businessRepository               repositories.BusinessRepository
	businessChannelSettingRepository repositories.BusinessChannelSettingsRepository
}

func NewBusinessChannelSettings(br repositories.BusinessRepository, bsr repositories.BusinessChannelSettingsRepository) BusinessChannelSetting {
	return &businessChannelSetting{
		businessRepository:               br,
		businessChannelSettingRepository: bsr,
	}
}

func (s *businessChannelSetting) Get(ctx context.Context, businessID, userID uuid.UUID) (*model.BusinessChannelSetting, error) {
	if _, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID); err != nil {
		return nil, err
	}

	return s.businessChannelSettingRepository.FindByBusinessID(ctx, businessID)
}

func (s *businessChannelSetting) Delete(ctx context.Context, businessID, userID uuid.UUID) error {
	if _, err := s.businessChannelSettingRepository.FindByBusinessID(ctx, businessID); err != nil {
		return err
	}

	return s.businessChannelSettingRepository.DeleteByBusinessID(ctx, businessID)
}
