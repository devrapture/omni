package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/devrapture/omni/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileUploadHandler struct{}

func NewFileUploadHandler() *FileUploadHandler {
	return &FileUploadHandler{}
}

func (h *FileUploadHandler) HandleFileUpload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)
	if err := c.Request.ParseMultipartForm(5 << 20); err != nil {
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

	head := make([]byte, 512)
	n, _ := src.Read(head)
	mime := http.DetectContentType(head[:n])

	allowedMIME := map[string]bool{
		"application/pdf": true,
		"text/csv":        true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}
	if !allowedMIME[mime] {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "unsupported file content type")
		return
	}

	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to create uploads directory")
		return
	}

	storedFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dist := filepath.Join("./uploads", storedFileName)
	if err := c.SaveUploadedFile(file, dist); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to store file")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "file uploaded successfully", nil, nil)
}
