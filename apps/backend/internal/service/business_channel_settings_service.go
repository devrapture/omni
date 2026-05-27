package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/devrapture/omni/internal/config"
	"github.com/devrapture/omni/internal/dto"
	apperrors "github.com/devrapture/omni/internal/errors"
	"github.com/devrapture/omni/internal/integrations/telegram"
	"github.com/devrapture/omni/internal/model"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/devrapture/omni/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BusinessChannelSetting interface {
	Get(ctx context.Context, businessID, userID uuid.UUID) (*dto.UpdateBusinessChannelResponse, error)
	Update(ctx context.Context, businessID, userID uuid.UUID, req dto.UpdateBusinessChannelSettingDTO) (*model.BusinessChannelSetting, error)
}

type businessChannelSetting struct {
	businessRepository               repositories.BusinessRepository
	businessChannelSettingRepository repositories.BusinessChannelSettingsRepository
	logger                           *zap.Logger
	telegramClient                   *telegram.TelegramClient
	cfg                              *config.Config
}

func NewBusinessChannelSettings(br repositories.BusinessRepository, bsr repositories.BusinessChannelSettingsRepository, logger *zap.Logger, cfg *config.Config) BusinessChannelSetting {
	return &businessChannelSetting{
		businessRepository:               br,
		businessChannelSettingRepository: bsr,
		logger:                           logger,
		telegramClient:                   telegram.NewTelegramClient(cfg),
		cfg:                              cfg,
	}
}

func (s *businessChannelSetting) Get(ctx context.Context, businessID, userID uuid.UUID) (*dto.UpdateBusinessChannelResponse, error) {
	if _, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID); err != nil {
		return nil, err
	}

	setting, err := s.businessChannelSettingRepository.FindByBusinessID(ctx, businessID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &dto.UpdateBusinessChannelResponse{
			BusinessID:     businessID,
			TelegramActive: false,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	response := &dto.UpdateBusinessChannelResponse{
		BusinessID:     setting.BusinessID,
		TelegramActive: setting.TelegramActive,
	}
	if setting.TelegramBotUsername != nil {
		response.TelegramUserName = *setting.TelegramBotUsername
	}

	return response, nil
}

func (s *businessChannelSetting) Update(ctx context.Context, businessID, userID uuid.UUID, req dto.UpdateBusinessChannelSettingDTO) (*model.BusinessChannelSetting, error) {
	if _, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID); err != nil {
		return nil, err
	}

	telegramBotToken := strings.TrimSpace(req.TelegramBotToken)
	if telegramBotToken == "" {
		return nil, apperrors.ErrInvalidTelegramBotToken
	}

	validationCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	username, err := s.telegramClient.ValidateBotToken(validationCtx, telegramBotToken)
	if err != nil {
		return nil, apperrors.ErrInvalidTelegramBotToken
	}
	encryptedToken, err := utils.EncryptText(telegramBotToken, s.cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}
	setting := &model.BusinessChannelSetting{
		BusinessID:                businessID,
		TelegramBotUsername:       &username,
		TelegramBotTokenEncrypted: &encryptedToken,
		TelegramActive:            true,
	}
	if err := s.businessChannelSettingRepository.Upsert(ctx, setting); err != nil {
		return nil, err
	}
	s.logger.Info("telegram bot token updated", zap.String("user_id", userID.String()), zap.String("business_id", businessID.String()), zap.String("telegram_bot_token", encryptedToken))
	return setting, nil
}

func (s *businessChannelSetting) Delete(ctx context.Context, businessID, userID uuid.UUID) error {
	if _, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID); err != nil {
		return err
	}

	return s.businessChannelSettingRepository.DeleteByBusinessID(ctx, businessID)
}
