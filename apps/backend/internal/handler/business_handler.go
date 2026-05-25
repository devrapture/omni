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

func (h *BusinessHandler) ListSources(c *gin.Context) {
	userID, _ := c.Get("userID")
	businessID, err := uuid.Parse(c.Param("businessID"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid business_id")
		return
	}
	// sourceTypeQuery := c.DefaultQuery("source_type", "all")
	var req dto.ListSourcesRequest

	if err := c.ShouldBind(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	if req.SourceType == "" {
		req.SourceType = "all"
	}

	sources, err := h.service.ListSources(c.Request.Context(), businessID, userID.(uuid.UUID), req.SourceType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "BUSINESS_NOT_FOUND", "business not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to get knowledge")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful retrieved sources", sources, nil)
}

func (h *BusinessHandler) DeleteSource(c *gin.Context) {
	userID, _ := c.Get("userID")
	businessID, err := uuid.Parse(c.Param("businessID"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid business_id")
		return
	}

	var req dto.DeleteSourceDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	if err := h.service.DeleteBySource(c.Request.Context(), businessID, userID.(uuid.UUID), req.SourceName); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to delete source")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful deleted source", nil, nil)
}

func (h *BusinessHandler) AddText(c *gin.Context) {
	userID, _ := c.Get("userID")
	businessID, err := uuid.Parse(c.Param("businessID"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid business_id")
		return
	}
	var req dto.AddTextDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	_, err = h.service.AddText(c.Request.Context(), businessID, userID.(uuid.UUID), req.Title, req.Content)
	if err != nil {
		log.Println("Add text error", err)
		switch {
		case errors.Is(err, apperrors.ErrMissingUserGeminiKey):
			utils.ErrorResponse(c, http.StatusBadRequest, "GEMINI_KEY_MISSING", "Please add your Gemini API key in settings.")
		case errors.Is(err, apperrors.ErrInvalidGeminiKey):
			utils.ErrorResponse(c, http.StatusBadRequest, "GEMINI_KEY_INVALID", "Your Gemini API key is invalid. Please update it in settings.")
		case errors.Is(err, apperrors.ErrGeminiQuotaExceeded):
			utils.ErrorResponse(c, http.StatusPaymentRequired, "GEMINI_KEY_QUOTA_EXCEEDED", "Your Gemini API key has exhausted its quota or rate limit.")
		case errors.Is(err, apperrors.ErrGeminiKeyRejected):
			utils.ErrorResponse(c, http.StatusBadGateway, "GEMINI_KEY_ERROR", "Gemini rejected your API key. Please check your key and billing settings.")
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to add text")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful added text", nil, nil)
}
