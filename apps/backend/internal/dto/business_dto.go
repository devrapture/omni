package dto

type CreateBusinessDTO struct {
	Name string `json:"name" binding:"required"`
}

type DeleteSourceDTO struct {
	SourceName string `json:"source_name" binding:"required"`
}

type AddTextDTO struct {
	Title string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}