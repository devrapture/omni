package dto

import "github.com/google/uuid"

type PresignUploadRequest struct {
	FileName string `json:"file_name" binding:"required"`
}

type CompleteUploadRequest struct {
	ObjectKey  string    `json:"object_key" binding:"required"`
	BusinessId uuid.UUID `json:"business_id" binding:"required,uuid4"`
}
