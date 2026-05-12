package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIProvider string

const (
	AIProviderGemini AIProvider = "gemini"
	AIProviderOpenAI AIProvider = "openai"
	// TODO: add more ai providers in future if needed
)

type AIKeyMode string

const (
	AIKeyModePlatform AIKeyMode = "platform"
	AIKeyModeUserKey  AIKeyMode = "user_key"
)

type UserSetting struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	User   User      `json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`

	Provider        AIProvider `json:"provider" gorm:"type:text;not null;default:gemini"`
	Mode            AIKeyMode  `json:"mode" gorm:"type:text;not null;default:platform"`
	APIKeyEncrypted string     `json:"-" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

func (u *UserSetting) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
