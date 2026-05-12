package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/devrapture/omni/internal/config"
	"github.com/devrapture/omni/internal/model"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/tasks"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const (
	maxFileNameLen = 255
)

type FileUploadHandler struct {
	cfg         *config.Config
	service     *service.ParserService
	asynqClient *asynq.Client
	logger      *zap.Logger
}

type fileUploadResponse struct {
	TaskID string                `json:"task_id"`
	Queue  string                `json:"queue"`
	Status model.UploadJobStatus `json:"status"`
}

func NewFileUploadHandler(cfg *config.Config, service *service.ParserService, asynqClient *asynq.Client, logger *zap.Logger) *FileUploadHandler {
	return &FileUploadHandler{
		cfg:         cfg,
		service:     service,
		asynqClient: asynqClient,
		logger:      logger,
	}
}

func (h *FileUploadHandler) HandleFileUpload(c *gin.Context) {
	userID, _ := c.Get("userID")
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

	if err := os.MkdirAll("./uploads", 0o755); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to create uploads directory")
		return
	}

	fileName := filepath.Base(file.Filename)

	if len(fileName) > maxFileNameLen {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "file name is too long")
		return
	}

	storedFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dist := filepath.Join("./uploads", storedFileName)
	if err := c.SaveUploadedFile(file, dist); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to store file")
		return
	}

	task, err := tasks.NewFileParseTask(userID.(uuid.UUID), dist)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to create file parse task")
		return
	}

	info, err := h.asynqClient.Enqueue(task, tasks.FileParseOptions()...)
	if err != nil {
		if removeErr := os.Remove(dist); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			h.logger.Warn("failed to cleanup uploaded file after enqueueing file parse task", zap.Error(removeErr))
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to enqueue file parse task")
		return
	}

	utils.SuccessResponse(c, http.StatusAccepted, "file uploaded and queued for parsing", fileUploadResponse{
		TaskID: info.ID,
		Queue:  info.Queue,
		Status: model.UploadJobProcessing,
	}, nil)
}
