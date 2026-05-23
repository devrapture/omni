package dto

type CreateBusinessDTO struct {
	Name string `json:"name" binding:"required"`
}

type DeleteSourceDTO struct {
	SourceName string `json:"source_name" binding:"required"`
}
