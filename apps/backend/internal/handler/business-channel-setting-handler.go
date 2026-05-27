package handler

import (
	"errors"
	"log"
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
	log.Println("uuid.Parse(c.Param(businessID))", err)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid business_id")
		return
	}

	var req dto.UpdateBusinessChannelSettingDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	channelSettings, err := h.service.Update(c.Request.Context(), businessID, userID.(uuid.UUID), req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "BUSINESS_NOT_FOUND", "business not found")
			return
		}
		if errors.Is(err, apperrors.ErrInvalidTelegramBotToken) {
			utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_TELEGRAM_BOT_TOKEN", "invalid telegram bot token")
			return
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful updated business channel setting", channelSettings, nil)
}
