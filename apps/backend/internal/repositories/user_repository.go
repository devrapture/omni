package repositories

import (
	"context"
	"errors"

	"github.com/devrapture/omni/internal/models"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

type UserRepository interface {
	FindOrCreateUser(ctx context.Context, userID, userEmail, userName, userPicture, provider string) (*models.User, error)
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindOrCreateUser(ctx context.Context, userID, userEmail, userName, userPicture, provider string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("email=?", userEmail).First(&user)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		user = models.User{
			Name:       userName,
			Email:      userEmail,
			ProviderID: userID,
			Provider:   provider,
			AvatarURL:  userPicture,
		}
		if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			user.UserSetting = &models.UserSetting{
				UserID:   user.ID,
				Provider: models.AIProviderGemini,
				Mode:     models.AIKeyModePlatform,
			}
			return tx.Create(user.UserSetting).Error
		}); err != nil {
			return nil, err
		}
	}

	return &user, nil
}
