package dto

import "github.com/google/uuid"

type UpdateBusinessChannelSettingDTO struct {
	TelegramBotToken string `json:"telegram_bot_token" binding:"omitempty"`
}

type UpdateBusinessChannelResponse struct {
	BusinessID       uuid.UUID `json:"business_id"`
	TelegramActive   bool      `json:"telegram_active"`
	TelegramUserName string    `json:"telegram_user_name,omitempty"`
}
