package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Business struct {
	ID        uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string              `json:"name" gorm:"type:text;not null"`

	UserID    uuid.UUID           `json:"user_id" gorm:"type:uuid;not null"`
	User      User                `json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	
	Knowledge []BusinessKnowledge `json:"knowledge,omitempty" gorm:"foreignKey:BusinessID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

func (b *Business) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
