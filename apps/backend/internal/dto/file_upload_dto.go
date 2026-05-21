package dto

type PresignUploadRequest struct {
	FileName string `json:"file_name" binding:"required"`
}

type CompleteUploadRequest struct {
	ObjectKey  string `json:"object_key" binding:"required"`
	BusinessId string `json:"business_id" binding:"required"`
}
