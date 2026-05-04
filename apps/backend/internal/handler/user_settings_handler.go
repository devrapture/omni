package handler

import (
	"net/http"

	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "USER_SETTINGS_FETCH_FAILED", "failed to get user settings")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful retrieved user settings", userSetting, nil)
}
