package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/devrapture/omni/internal/config"
	apperrors "github.com/devrapture/omni/internal/errors"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)


const (
	maxFileNameLen = 255
)

type FileUploadHandler struct {
	cfg     *config.Config
	service *service.ParserService
	logger  *zap.Logger
}

type fileUploadResponse struct {
	Content    string `json:"content"`
	SourceType string `json:"source_type"`
	FileName   string `json:"file_name"`
}

func NewFileUploadHandler(cfg *config.Config, service *service.ParserService, logger *zap.Logger) *FileUploadHandler {
	return &FileUploadHandler{
		cfg:     cfg,
		service: service,
		logger:  logger,
	}
}

func (h *FileUploadHandler) HandleFileUpload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.FileUploadMaxBytes)
	if err := c.Request.ParseMultipartForm(h.cfg.FileUploadMaxBytes); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "file size is too large")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "file is required")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	src, err := file.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "failed to read uploaded file")
		return
	}
	defer src.Close()

	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to create uploads directory")
		return
	}

	fileName := filepath.Base(file.Filename)

	if len(fileName) > maxFileNameLen {
		utils.ErrorResponse(c,http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "file name is too long")
		return
	}

	storedFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dist := filepath.Join("./uploads", storedFileName)
	if err := c.SaveUploadedFile(file, dist); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to store file")
		return
	}

	content, sourceType, err := h.service.Parse(dist)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotSupportFile) {
			utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "unsupported file content type")
			return
		}
		if errors.Is(err, apperrors.ErrFileNameTooLong) {
			utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "file name is too long")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to parse file")
		return
	}

	h.logger.Info("File parsed successfully", zap.String("source_type", sourceType), zap.Int("content_length", len(content)))

	utils.SuccessResponse(c, http.StatusOK, "file uploaded successfully", fileUploadResponse{
		Content:    content,
		SourceType: sourceType,
		FileName:   fileName,
	}, nil)
}
