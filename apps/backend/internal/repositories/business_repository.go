package repositories

import (
	"context"

	"github.com/devrapture/omni/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusinessRepository interface {
	CreateBusiness(ctx context.Context, businessName string, userID uuid.UUID) (*model.Business, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Business, error)
}

type businessRepository struct {
	db *gorm.DB
}

func NewBusinessRepository(DB *gorm.DB) BusinessRepository {
	return &businessRepository{
		db: DB,
	}
}

func (r *businessRepository) CreateBusiness(ctx context.Context, businessName string, userID uuid.UUID) (*model.Business, error) {
	business := &model.Business{
		Name:   businessName,
		UserID: userID,
	}
	if err := r.db.WithContext(ctx).Create(business).Error; err != nil {
		return nil, err
	}
	return business, nil
}

func (r *businessRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Business, error) {
	var business model.Business
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&business).Error
	if err != nil {
		return nil, err
	}

	return &business, nil
}
