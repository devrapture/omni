package handler

import (
	"errors"
	"net/http"

	"github.com/devrapture/omni/internal/dto"
	apperrors "github.com/devrapture/omni/internal/errors"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	service service.UserSettingService
}

func NewUserSettingsHandler(service service.UserSettingService) *SettingsHandler {
	return &SettingsHandler{
		service: service,
	}
}

func (h *SettingsHandler) GetUserSettings(c *gin.Context) {
	userID, _ := c.Get("userID")
	userSetting, err := h.service.GetUserSettings(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, apperrors.ErrSettingsNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "SETTINGS_NOT_FOUND", "user settings not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "USER_SETTINGS_FETCH_FAILED", "failed to get user settings")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful retrieved user settings", userSetting, nil)
}

func (h *SettingsHandler) UpdateUserSettings(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req dto.UpdateUserSettingsDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if err := h.service.UpdateUserSettings(c.Request.Context(), userID.(uuid.UUID), req); err != nil {
		if errors.Is(err, apperrors.ErrEmptyAPIKey) {
			utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "api_key is required when mode is user_key")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "USER_SETTINGS_UPDATE_FAILED", "failed to update user settings")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful updated user settings", nil, nil)
}

func (h *SettingsHandler) DeleteUserSettings(c *gin.Context) {
	userID, _ := c.Get("userID")
	err := h.service.DeleteUserKey(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "SETTINGS_NOT_FOUND", "user settings not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "USER_SETTINGS_DELETE_FAILED", "failed to delete user settings")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Successful deleted user settings", nil, nil)
}
