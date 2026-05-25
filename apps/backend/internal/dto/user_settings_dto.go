package dto

import "github.com/devrapture/omni/internal/model"

type UpdateUserSettingsDTO struct {
	Provider model.AIProvider `json:"provider" binding:"required,oneof=gemini"`
	Mode     model.AIKeyMode  `json:"mode" binding:"required,oneof=user_key platform"`
	APIKey   string           `json:"api_key"`
}
