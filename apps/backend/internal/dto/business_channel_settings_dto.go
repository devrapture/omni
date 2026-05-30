package dto

import "github.com/google/uuid"

type UpdateBusinessChannelSettingDTO struct {
	TelegramBotToken *string `json:"telegram_bot_token" binding:"omitempty"`
	TelegramActive   *bool   `json:"telegram_active" binding:"omitempty"`
}

type UpdateBusinessChannelResponse struct {
	BusinessID       uuid.UUID `json:"business_id"`
	TelegramActive   bool      `json:"telegram_active"`
	TelegramUserName string    `json:"telegram_user_name,omitempty"`
}

type TelegramWebhookOutcome string

const (
	TelegramWebhookOutcomeNone               TelegramWebhookOutcome = ""
	TelegramWebhookOutcomeRegistered         TelegramWebhookOutcome = "registered"
	TelegramWebhookOutcomeRegistrationFailed TelegramWebhookOutcome = "registration_failed"
	TelegramWebhookOutcomeDeleted            TelegramWebhookOutcome = "deleted"
	TelegramWebhookOutcomeDeletionFailed     TelegramWebhookOutcome = "deletion_failed"
)

type BusinessChannelSettingUpdateResult struct {
	Setting        UpdateBusinessChannelResponse `json:"-"`
	WebhookOutcome TelegramWebhookOutcome        `json:"-"`
}
