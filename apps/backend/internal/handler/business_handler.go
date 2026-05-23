package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/devrapture/omni/internal/dto"
	"github.com/devrapture/omni/internal/model"
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
	entries, err := h.service.GetKnowledgeForUser(c.Request.Context(), businessID, userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "BUSINESS_NOT_FOUND", "business not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to get knowledge")
		return
	}

	type SourceSummary struct {
		SourceName string           `json:"source_name"`
		SourceType model.SourceType `json:"source_type"`
		ChunkCount int              `json:"chunk_count"`
		CreatedAt  time.Time        `json:"created_at"`
	}

	sourceMap := make(map[string]*SourceSummary)
	for _, e := range entries {
		key := e.SourceName
		if _, exists := sourceMap[key]; !exists {
			sourceMap[key] = &SourceSummary{
				SourceName: e.SourceName,
				SourceType: e.SourceType,
				CreatedAt:  e.CreatedAt,
			}
		}
		sourceMap[key].ChunkCount++
	}

	var sources []SourceSummary
	for _, s := range sourceMap {
		sources = append(sources, *s)
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to add text")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successful added text", nil, nil)
}
