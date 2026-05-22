package handler

import (
	"net/http"

	"github.com/devrapture/omni/internal/dto"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BusinessHandler struct {
	service service.BusinessService
}

func NewBusinessHandler(service service.BusinessService) *BusinessHandler {
	return &BusinessHandler{
		service: service,
	}
}

func (h *BusinessHandler) CreateBusiness(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req dto.CreateBusinessDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	business, err := h.service.CreateBusiness(c.Request.Context(), req.Name, userID.(uuid.UUID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "BUSINESS_CREATE_FAILED", "failed to create business")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful created business", business, nil)

}
