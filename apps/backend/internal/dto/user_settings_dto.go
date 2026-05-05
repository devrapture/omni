package dto

import "github.com/devrapture/omni/internal/models"

type UpdateUserSettingsDTO struct {
	Provider models.AIProvider `json:"provider" binding:"required,oneof=gemini"`
	Mode     models.AIKeyMode  `json:"mode" binding:"required,oneof=user_key platform"`
	APIKey   string            `json:"api_key"`
}
