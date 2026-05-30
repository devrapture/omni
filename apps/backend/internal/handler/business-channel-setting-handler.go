package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/devrapture/omni/internal/dto"
	apperrors "github.com/devrapture/omni/internal/errors"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusinessChannelSettingHandler struct {
	service service.BusinessChannelSetting
}

func NewBusinessChannelSettingHandler(service service.BusinessChannelSetting) *BusinessChannelSettingHandler {
	return &BusinessChannelSettingHandler{
		service: service,
	}
}

func (h *BusinessChannelSettingHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	businessID, err := uuid.Parse(c.Param("businessID"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid business_id")
		return
	}

	setting, err := h.service.Get(c.Request.Context(), businessID, userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "BUSINESS_NOT_FOUND", "business not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "BUSINESS_CHANNEL_SETTING_GET_FAILED", "failed to get business channel setting")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful retrieved business channel setting", setting, nil)
}

func (h *BusinessChannelSettingHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	businessID, err := uuid.Parse(c.Param("businessID"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid business_id")
		return
	}

	var req dto.UpdateBusinessChannelSettingDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	result, err := h.service.Update(c.Request.Context(), businessID, userID.(uuid.UUID), req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "BUSINESS_NOT_FOUND", "business not found")
			return
		}
		if errors.Is(err, apperrors.ErrInvalidTelegramBotFormat) {
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_TELEGRAM_BOT_TOKEN", apperrors.ErrInvalidTelegramBotFormat.Error())
			return
		}
		if errors.Is(err, apperrors.ErrInvalidTelegramBotToken) {
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_TELEGRAM_BOT_TOKEN", "invalid telegram bot token")
			return
		}
		if errors.Is(err, apperrors.ErrTelegramBotTokenNotProvided) {
			utils.ErrorResponse(c, http.StatusBadRequest, "TELEGRAM_BOT_TOKEN_REQUIRED", "telegram bot token is not provided")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "BUSINESS_CHANNEL_SETTING_UPDATE_FAILED", "failed to update business channel setting")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, updateSuccessMessage(result), result.Setting, nil)
}

func updateSuccessMessage(result *dto.BusinessChannelSettingUpdateResult) string {
	switch result.WebhookOutcome {
	case dto.TelegramWebhookOutcomeRegistered:
		if result.Setting.TelegramUserName != "" {
			return fmt.Sprintf(
				"Your Telegram bot @%s is connected. Customers can message your bot and receive AI-powered replies.",
				result.Setting.TelegramUserName,
			)
		}
		return "Your Telegram bot is connected. Customers can message your bot and receive AI-powered replies."
	case dto.TelegramWebhookOutcomeRegistrationFailed:
		return "Your bot token was saved, but we couldn't register the webhook to receive messages. Please try again."
	case dto.TelegramWebhookOutcomeDeleted:
		return "Telegram channel has been deactivated."
	case dto.TelegramWebhookOutcomeDeletionFailed:
		return "Settings saved, but we couldn't remove the Telegram webhook. Your bot may still receive messages until this is resolved."
	default:
		return "Business channel settings updated successfully."
	}
}
