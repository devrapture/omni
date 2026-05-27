package handler

import (
	"net/http"

	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "BUSINESS_CHANNEL_SETTING_GET_FAILED", "failed to get business channel setting")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful retrieved business channel setting", setting, nil)
}
