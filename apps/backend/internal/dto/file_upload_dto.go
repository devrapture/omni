package dto

type PresignUploadRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

type CompleteUploadRequest struct {
	ObjectKey string `json:"object_key" binding:"required"`
}
