package repositories

import (
	"context"
	"errors"

	"github.com/devrapture/omni/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSettingRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserSetting, error)
	Upsert(ctx context.Context, userID uuid.UUID, userSettings *models.UserSetting) error
	DeleteAPIKey(ctx context.Context, userID uuid.UUID) error
}

type userSettingRepository struct {
	db *gorm.DB
}

func NewUserSettingRepository(DB *gorm.DB) UserSettingRepository {
	return &userSettingRepository{
		db: DB,
	}
}

func (r *userSettingRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserSetting, error) {
	var setting models.UserSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&setting).Error
	return &setting, err
}

func (r *userSettingRepository) Upsert(ctx context.Context, userID uuid.UUID, userSettings *models.UserSetting) error {
	var existing models.UserSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&existing).Error
	if err != nil {
		existing.Mode = userSettings.Mode
		if userSettings.GeminiAPIKeyEncrypted != "" {
			existing.GeminiAPIKeyEncrypted = userSettings.GeminiAPIKeyEncrypted
		}
		return r.db.WithContext(ctx).Save(&existing).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return r.db.WithContext(ctx).Create(userSettings).Error
}

func (r *userSettingRepository) DeleteAPIKey(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.UserSetting{}).Where("user_id = ?", userID).Updates(
		map[string]interface{}{
			"mode":                     models.AIKeyModePlatform,
			"gemini_api_key_encrypted": "",
		},
	).Error
}
