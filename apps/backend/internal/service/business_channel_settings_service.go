package service

import (
	"context"
	"errors"
	"fmt"
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
	Update(ctx context.Context, businessID, userID uuid.UUID, req dto.UpdateBusinessChannelSettingDTO) (*dto.BusinessChannelSettingUpdateResult, error)
	Delete(ctx context.Context, businessID, userID uuid.UUID, webHookURL string) error
	GetDecryptedTelegramBotToken(ctx context.Context, businessID, userID uuid.UUID) (string, error)
	RegisterAllTelegramWebhooks(ctx context.Context, baseURL string) error
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

func (s *businessChannelSetting) Update(ctx context.Context, businessID, userID uuid.UUID, req dto.UpdateBusinessChannelSettingDTO) (*dto.BusinessChannelSettingUpdateResult, error) {
	if _, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID); err != nil {
		return nil, err
	}

	existingSetting, err := s.businessChannelSettingRepository.FindByBusinessID(ctx, businessID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hasNewToken := req.TelegramBotToken != nil
	hasActive := req.TelegramActive != nil

	var telegramActive bool
	var username string
	var encryptedToken string

	if hasNewToken {
		telegramBotToken := strings.TrimSpace(*req.TelegramBotToken)
		if telegramBotToken == "" {
			return nil, apperrors.ErrInvalidTelegramBotToken
		}

		// Basic format validation for Telegram tokens: digits:alphanumeric
		if !s.isValidTelegramToken(telegramBotToken) {
			return nil, apperrors.ErrInvalidTelegramBotFormat
		}

		validationCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		username, err = s.telegramClient.ValidateBotToken(validationCtx, telegramBotToken)
		if err != nil {
			return nil, apperrors.ErrInvalidTelegramBotToken
		}
		encryptedToken, err = utils.EncryptText(telegramBotToken, s.cfg.EncryptionKey)
		if err != nil {
			return nil, err
		}

		if hasActive {
			telegramActive = *req.TelegramActive
		} else {
			telegramActive = true
		}
	} else {
		hasExistingToken := existingSetting != nil && existingSetting.TelegramBotTokenEncrypted != nil && *existingSetting.TelegramBotTokenEncrypted != ""

		if hasActive {
			if !hasExistingToken {
				return nil, apperrors.ErrTelegramBotTokenNotProvided
			}
			telegramActive = *req.TelegramActive
			encryptedToken = *existingSetting.TelegramBotTokenEncrypted
			if existingSetting.TelegramBotUsername != nil {
				username = *existingSetting.TelegramBotUsername
			}
		} else {
			if !hasExistingToken {
				return nil, apperrors.ErrTelegramBotTokenNotProvided
			}
			telegramActive = existingSetting.TelegramActive
			encryptedToken = *existingSetting.TelegramBotTokenEncrypted
			if existingSetting.TelegramBotUsername != nil {
				username = *existingSetting.TelegramBotUsername
			}
		}
	}

	var id uuid.UUID
	if existingSetting != nil {
		id = existingSetting.ID
	}

	setting := &model.BusinessChannelSetting{
		ID:                        id,
		BusinessID:                businessID,
		TelegramBotUsername:       &username,
		TelegramBotTokenEncrypted: &encryptedToken,
		TelegramActive:            telegramActive,
	}

	if err := s.businessChannelSettingRepository.Upsert(ctx, setting); err != nil {
		return nil, err
	}

	tokenChanged := hasNewToken && (existingSetting == nil ||
		existingSetting.TelegramBotTokenEncrypted == nil ||
		*existingSetting.TelegramBotTokenEncrypted != encryptedToken)
	webhookOutcome := s.syncTelegramWebhookAfterUpdate(ctx, businessID, userID, existingSetting, setting, tokenChanged)

	s.logger.Info("telegram bot token updated/toggled", zap.String("user_id", userID.String()), zap.String("business_id", businessID.String()))
	return &dto.BusinessChannelSettingUpdateResult{
		Setting:        toBusinessChannelResponse(setting),
		WebhookOutcome: webhookOutcome,
	}, nil
}

func toBusinessChannelResponse(setting *model.BusinessChannelSetting) dto.UpdateBusinessChannelResponse {
	response := dto.UpdateBusinessChannelResponse{
		BusinessID:     setting.BusinessID,
		TelegramActive: setting.TelegramActive,
	}
	if setting.TelegramBotUsername != nil {
		response.TelegramUserName = *setting.TelegramBotUsername
	}
	return response
}

func (s *businessChannelSetting) syncTelegramWebhookAfterUpdate(
	ctx context.Context,
	businessID, userID uuid.UUID,
	existingSetting, setting *model.BusinessChannelSetting,
	tokenChanged bool,
) dto.TelegramWebhookOutcome {
	webhookURL := fmt.Sprintf("%s/api/v1/webhooks/telegram/%s", s.cfg.AppBaseUrl, businessID)

	if tokenChanged && existingSetting != nil &&
		existingSetting.TelegramBotTokenEncrypted != nil &&
		*existingSetting.TelegramBotTokenEncrypted != "" {
		oldToken, err := utils.DecryptText(*existingSetting.TelegramBotTokenEncrypted, s.cfg.EncryptionKey)
		if err != nil {
			s.logger.Warn("failed to decrypt previous telegram bot token for webhook cleanup", zap.Error(err))
		} else if err := s.telegramClient.DeleteWebHook(ctx, oldToken, webhookURL); err != nil {
			s.logger.Warn("failed to delete webhook for previous telegram bot token", zap.Error(err))
		}
	}

	if !setting.TelegramActive {
		if setting.TelegramBotTokenEncrypted == nil || *setting.TelegramBotTokenEncrypted == "" {
			return dto.TelegramWebhookOutcomeNone
		}
		wasPreviouslyActive := existingSetting != nil && existingSetting.TelegramActive
		if err := s.deleteTelegramWebhook(ctx, businessID, userID, webhookURL); err != nil {
			s.logger.Warn("failed to delete telegram webhook after deactivation", zap.Error(err))
			if wasPreviouslyActive {
				return dto.TelegramWebhookOutcomeDeletionFailed
			}
			return dto.TelegramWebhookOutcomeNone
		}
		if wasPreviouslyActive {
			return dto.TelegramWebhookOutcomeDeleted
		}
		return dto.TelegramWebhookOutcomeNone
	}

	if setting.TelegramBotTokenEncrypted == nil || *setting.TelegramBotTokenEncrypted == "" {
		return dto.TelegramWebhookOutcomeNone
	}

	if err := s.registerTelegramWebhook(ctx, businessID, userID, webhookURL); err != nil {
		s.logger.Warn("failed to register telegram webhook", zap.Error(err))
		return dto.TelegramWebhookOutcomeRegistrationFailed
	}
	return dto.TelegramWebhookOutcomeRegistered
}

func (s *businessChannelSetting) Delete(ctx context.Context, businessID, userID uuid.UUID, webHookURL string) error {
	if _, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID); err != nil {
		return err
	}

	existingSetting, err := s.businessChannelSettingRepository.FindByBusinessID(ctx, businessID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existingSetting != nil && existingSetting.TelegramBotTokenEncrypted != nil && *existingSetting.TelegramBotTokenEncrypted != "" {
		webhookURL := fmt.Sprintf("%s/api/v1/webhooks/telegram/%s", s.cfg.AppBaseUrl, businessID)
		if err := s.deleteTelegramWebhook(ctx, businessID, userID, webhookURL); err != nil {
			s.logger.Warn("Failed to delete telegram webhook during settings deletion", zap.Error(err))
		}
	}

	return s.businessChannelSettingRepository.DeleteByBusinessID(ctx, businessID)
}

func (s *businessChannelSetting) isValidTelegramToken(token string) bool {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) < 5 || len(parts[1]) < 20 {
		return false
	}
	// First part should be all digits
	for _, r := range parts[0] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (s *businessChannelSetting) registerTelegramWebhook(ctx context.Context, businessID, userID uuid.UUID, webhookURL string) error {
	decryptedTelegramBotToken, err := s.GetDecryptedTelegramBotToken(ctx, businessID, userID)
	if err != nil {
		return err
	}

	s.logger.Info("Registering telegram webhook", zap.String("business_id", businessID.String()), zap.String("webhook_url", webhookURL))
	err = s.telegramClient.SetWebHook(ctx, decryptedTelegramBotToken, webhookURL)
	if err != nil {
		return err
	}

	s.logger.Info("Telegram webhook registered successfully",
		zap.String("business_id", businessID.String()),
		zap.String("webhook_url", webhookURL),
	)
	return nil
}

// RegisterAllTelegramWebhooks finds every active Telegram config and registers
// (or re-registers) the webhook for each one.
//
// This is called at server startup. It handles:
// - First-time registration after a fresh deploy
// - Re-registration after the server URL changes (new domain, new ngrok URL)
// - Recovery after a previous registration failure
//
// It logs warnings for failures but does not abort — one bad token should not
// prevent other businesses' bots from working.

func (s *businessChannelSetting) RegisterAllTelegramWebhooks(ctx context.Context, baseURL string) error {
	s.logger.Info("Registering webhooks for all active Telegram bots",
		zap.String("base_url", baseURL),
	)
	allActiveSettings, err := s.businessChannelSettingRepository.FindAllActiveTelegramBot(ctx)
	if err != nil {
		return fmt.Errorf("failed to load active channel configs: %w", err)
	}

	if len(allActiveSettings) == 0 {
		s.logger.Info("No active Telegram bots found — skipping webhook registration")
		return nil
	}

	s.logger.Info("Found active Telegram bots", zap.Int("count", len(allActiveSettings)))

	successCount := 0
	failCount := 0
	for _, setting := range allActiveSettings {
		business, err := s.businessRepository.FindByID(ctx, setting.BusinessID)
		if err != nil {
			s.logger.Warn("Could  not find business for setting", zap.String("business_id", setting.BusinessID.String()), zap.Error(err))
			failCount++
			continue
		}
		webhookURL := fmt.Sprintf("%s/api/v1/webhooks/telegram/%s", baseURL, business.ID)
		decryptedTelegramBotToken, err := utils.DecryptText(*setting.TelegramBotTokenEncrypted, s.cfg.EncryptionKey)
		if err != nil {
			s.logger.Warn("Failed to decrypt token for business",
				zap.String("business_id", setting.BusinessID.String()),
				zap.Error(err),
			)
			failCount++
			continue
		}

		webhookInfo, err := s.telegramClient.GetWebhookInfo(ctx, decryptedTelegramBotToken)
		if err == nil && webhookInfo.URL == baseURL {
			s.logger.Info("Webhook already registered — skipping",
				zap.String("business_id", business.ID.String()),
				zap.String("url", webhookURL),
			)
			successCount++
			continue
		}

		// Register (or re-register) the webhook
		if err := s.telegramClient.SetWebHook(ctx, decryptedTelegramBotToken, webhookURL); err != nil {
			s.logger.Warn("Failed to register webhook",
				zap.String("business_id", business.ID.String()),
				zap.String("url", webhookURL),
				zap.Error(err),
			)
			failCount++
			continue
		}

		s.logger.Info("Webhook registered",
			zap.String("business_id", business.ID.String()),
			zap.String("url", webhookURL),
		)
		successCount++
	}

	s.logger.Info("Webhook registration complete",
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
	)

	return nil
}

func (s *businessChannelSetting) deleteTelegramWebhook(ctx context.Context, businessID, userID uuid.UUID, webhookURL string) error {
	decryptedTelegramBotToken, err := s.GetDecryptedTelegramBotToken(ctx, businessID, userID)
	if err != nil {
		return err
	}
	err = s.telegramClient.DeleteWebHook(ctx, decryptedTelegramBotToken, webhookURL)
	if err != nil {
		return err
	}
	s.logger.Info("Deleting Telegram webhook successful",
		zap.String("business_id", businessID.String()),
		zap.String("webhook_url", webhookURL),
	)
	return nil
}

func (s *businessChannelSetting) GetDecryptedTelegramBotToken(ctx context.Context, businessID, userID uuid.UUID) (string, error) {
	if _, err := s.businessRepository.FindByIDAndUserID(ctx, businessID, userID); err != nil {
		return "", err
	}

	existingSetting, err := s.businessChannelSettingRepository.FindByBusinessID(ctx, businessID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperrors.ErrTelegramBotTokenNotProvided
		}
		return "", err
	}

	if existingSetting == nil || existingSetting.TelegramBotTokenEncrypted == nil || *existingSetting.TelegramBotTokenEncrypted == "" {
		return "", apperrors.ErrTelegramBotTokenNotProvided
	}
	return utils.DecryptText(*existingSetting.TelegramBotTokenEncrypted, s.cfg.EncryptionKey)
}
