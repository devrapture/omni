package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email             string              `json:"email" gorm:"uniqueIndex:idx_users_email;not null" example:"john.doe@example.com" description:"User's email"`
	Name              string              `json:"name" gorm:"type:text;not null" example:"John Doe" description:"User's name"`
	AvatarURL         string              `json:"avatar_url" gorm:"type:text;not null" example:"https://example.com/avatar.png" description:"User's avatar URL"`
	Provider          string              `json:"provider" gorm:"type:text;not null" example:"google" description:"User's provider"`
	ProviderID        string              `json:"provider_id" gorm:"type:text;not null" example:"1234567890" description:"User's provider ID"`
	
	UserSetting       *UserSetting        `json:"user_setting" gorm:"omitempty"`
	
	Business          []Business          `json:"business,omitempty" gorm:"foreignKey:UserID"`
	
	CreatedAt         time.Time           `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt         time.Time           `json:"updated_at" example:"2021-01-01T00:00:00Z"`
	DeletedAt         gorm.DeletedAt      `json:"-" gorm:"index" example:"2021-01-01T00:00:00Z"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
