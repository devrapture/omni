package dto

type CreateBusinessDTO struct {
	Name string `json:"name" binding:"required"`
}
