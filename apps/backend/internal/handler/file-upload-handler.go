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
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "file is required")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".pdf":  true,
		".docx": true,
		".csv":  true,
		".xlsx": true,
	}

	if !allowed[ext] {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "unsupported file type")
		return
	}

	if file.Size > 5<<20 {
		utils.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "file size is too large")
		return
	}

	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "failed to create uploads directory")
		return
	}

	storedFileName := fmt.Sprintf("%s.%s", uuid.New().String(), ext)
	dist := filepath.Join("./uploads", storedFileName)
	c.SaveUploadedFile(file, dist)

	utils.SuccessResponse(c, http.StatusOK, "file uploaded and successfully parsed", nil, nil)
}
