package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusinessChannelSetting struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`

	BusinessID uuid.UUID `json:"business_id" gorm:"type:uuid;not null;uniqueIndex"`
	Business   Business  `json:"-" gorm:"foreignKey:BusinessID;references:ID;constraint:OnDelete:CASCADE"`

	TelegramBotTokenEncrypted *string `json:"-" gorm:"type:text"`
	TelegramBotUsername       *string `json:"telegram_bot_username,omitempty" gorm:"type:text"`
	TelegramActive            bool    `json:"telegram_active" gorm:"type:bool;default:true"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

func (b *BusinessChannelSetting) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
