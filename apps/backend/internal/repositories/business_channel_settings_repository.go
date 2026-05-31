package repositories

import (
	"context"

	"github.com/devrapture/omni/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BusinessChannelSettingsRepository interface {
	FindByBusinessID(ctx context.Context, businessID uuid.UUID) (*model.BusinessChannelSetting, error)
	Upsert(ctx context.Context, setting *model.BusinessChannelSetting) error
	DeleteByBusinessID(ctx context.Context, businessID uuid.UUID) error
	FindAllActiveTelegramBot(ctx context.Context) ([]model.BusinessChannelSetting, error)
}

type businessChannelSettingsRepository struct {
	db *gorm.DB
}

func NewBusinessChannelSettingsRepository(DB *gorm.DB) BusinessChannelSettingsRepository {
	return &businessChannelSettingsRepository{
		db: DB,
	}
}

func (r *businessChannelSettingsRepository) FindByBusinessID(ctx context.Context, businessID uuid.UUID) (*model.BusinessChannelSetting, error) {
	var setting model.BusinessChannelSetting
	err := r.db.WithContext(ctx).Where("business_id = ?", businessID).First(&setting).Error
	if err != nil {
		return nil, err
	}

	return &setting, nil
}

func (r *businessChannelSettingsRepository) Upsert(ctx context.Context, setting *model.BusinessChannelSetting) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "business_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"telegram_bot_token_encrypted",
			"telegram_bot_username",
			"telegram_active",
		}),
	}).Create(setting).Error
}

func (r *businessChannelSettingsRepository) DeleteByBusinessID(ctx context.Context, businessID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("business_id = ?", businessID).Delete(&model.BusinessChannelSetting{}).Error
}

func (r *businessChannelSettingsRepository) FindAllActiveTelegramBot(ctx context.Context) ([]model.BusinessChannelSetting, error) {
	var setting []model.BusinessChannelSetting
	err := r.db.WithContext(ctx).Where("telegram_active = true").Find(&setting).Error
	return setting, err
}
