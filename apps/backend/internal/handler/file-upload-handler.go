package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	// "os"
	"path/filepath"

	"github.com/devrapture/omni/internal/config"
	"github.com/devrapture/omni/internal/dto"
	"github.com/devrapture/omni/internal/model"
	"github.com/devrapture/omni/internal/repositories"
	"github.com/devrapture/omni/internal/service"
	"github.com/devrapture/omni/internal/storage"
	"github.com/devrapture/omni/internal/tasks"

	// "github.com/devrapture/omni/internal/tasks"
	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FileUploadHandler struct {
	cfg           *config.Config
	service       *service.ParserService
	asynqClient   *asynq.Client
	uploadJobRepo repositories.UploadJobRepository
	r2            *storage.R2Storage
	logger        *zap.Logger
}

func NewFileUploadHandler(cfg *config.Config, service *service.ParserService, asynqClient *asynq.Client, uploadJobRepo repositories.UploadJobRepository, r2 *storage.R2Storage, logger *zap.Logger) *FileUploadHandler {
	return &FileUploadHandler{
		cfg:           cfg,
		service:       service,
		asynqClient:   asynqClient,
		uploadJobRepo: uploadJobRepo,
		r2:            r2,
		logger:        logger,
	}
}


func (h *FileUploadHandler) GetUploadJob(c *gin.Context) {
	userID, _ := c.Get("userID")
	jobID, err := uuid.Parse(c.Param("jobID"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid job_id")
		return
	}

	job, err := h.uploadJobRepo.FindByIDandUserID(c.Request.Context(), userID.(uuid.UUID), jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "UPLOAD_JOB_NOT_FOUND", "upload job not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to get upload job")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "upload job retrieved", gin.H{
		"job_id":      job.ID,
		"status":      job.Status,
		"source_type": job.SourceType,
		"content":      job.Content,
		"error":        job.Error,
		"created_at":   job.CreatedAt,
		"updated_at":   job.UpdatedAt,
		"is_completed": job.Status == model.UploadJobCompleted,
		"is_failed":    job.Status == model.UploadJobFailed,
	}, nil)
}

func (h *FileUploadHandler) CreatePresignedUploadURL(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req dto.PresignUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	ext := filepath.Ext(req.FileName)
	if !isAllowedUploadExtension(ext) {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "file extension is not allowed")
		return
	}

	jobID := uuid.New()
	objectKey := fmt.Sprintf("uploads/%s/%s%s", userID.(uuid.UUID).String(), jobID.String(), ext)

	job := &model.UploadJob{
		ID:         jobID,
		Status:     model.UploadJobQueued,
		UserID:     userID.(uuid.UUID),
		ObjectKey:  objectKey,
		SourceType: ext,
	}

	if err := h.uploadJobRepo.Create(c.Request.Context(), job); err != nil {
		h.logger.Error("failed to create upload job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to create upload job")
		return
	}

	uploadURL, err := h.r2.PresignPutObject(c.Request.Context(), objectKey, time.Duration(h.cfg.R2_PRESIGN_TTL_SECONDS)*time.Second)
	if err != nil {
		h.logger.Error("failed to presign upload URL", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to presign upload URL")
		return
	}

	utils.SuccessResponse(c, http.StatusAccepted, "upload url created", gin.H{
		"job_id":     job.ID,
		"object_key": objectKey,
		"upload_url": uploadURL,
	}, nil)
}

func (h *FileUploadHandler) CompleteUpload(c *gin.Context) {
	userID, _ := c.Get("userID")
	jobID, err := uuid.Parse(c.Param("jobID"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "invalid job_id")
		return
	}

	var req dto.CompleteUploadRequest

	if err := c.ShouldBind(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	job, err := h.uploadJobRepo.FindByIDandUserID(c.Request.Context(), userID.(uuid.UUID), jobID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "UPLOAD_JOB_NOT_FOUND", "upload job not found")
		return
	}

	if job.ObjectKey != req.ObjectKey {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "object key does not match upload job")
		return
	}

	task, err := tasks.NewFileParseTask(job.ID, userID.(uuid.UUID), job.ObjectKey)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to create file parse task")
		return
	}

	info, err := h.asynqClient.Enqueue(task, tasks.FileParseOptions()...)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to enqueue file parse task")
		return
	}

	utils.SuccessResponse(c, http.StatusAccepted, "file uploaded and queued for parsing", gin.H{
		"job_id": job.ID.String(),
		"queue":  info.Queue,
		"status": job.Status,
	}, nil)

}

func isAllowedUploadExtension(ext string) bool {
	switch ext {
	case ".csv", ".pdf", ".docx", ".xlsx":
		return true
	default:
		return false
	}
}
