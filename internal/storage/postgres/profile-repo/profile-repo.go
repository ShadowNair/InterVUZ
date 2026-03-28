package profile_repo

import (
	"context"
	"errors"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (repo *ProfileRepository) CreateProfile(ctx context.Context, profile domain.Profile) (*domain.Profile, error) {
	model := ToProfileModel(profile)

	if err := repo.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}

	result := ToDomainProfile(model)
	return &result, nil
}

func (repo *ProfileRepository) GetProfileByEmail(ctx context.Context, email string) (*domain.Profile, error) {
	var model ProfileModel

	err := repo.db.WithContext(ctx).
		Where("email = ?", email).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	result := ToDomainProfile(model)
	return &result, nil
}

func (repo *ProfileRepository) GetProfileByID(ctx context.Context, id string) (*domain.Profile, error) {
	var model ProfileModel

	err := repo.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	result := ToDomainProfile(model)
	return &result, nil
}

func (repo *ProfileRepository) UpdateLastLoginAt(ctx context.Context, profileID string) error {
	now := time.Now().UTC()

	result := repo.db.WithContext(ctx).
		Model(&ProfileModel{}).
		Where("id = ?", profileID).
		Updates(map[string]any{
			"last_login_at": now,
			"updated_at":    now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *ProfileRepository) CreateRefreshToken(ctx context.Context, token domain.RefreshToken) (*domain.RefreshToken, error) {
	model := ToRefreshTokenModel(token)

	if err := repo.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}

	result := ToDomainRefreshToken(model)
	return &result, nil
}

func (repo *ProfileRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var model RefreshTokenModel

	err := repo.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	result := ToDomainRefreshToken(model)
	return &result, nil
}

func (repo *ProfileRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	now := time.Now().UTC()

	result := repo.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("token_hash = ?", tokenHash).
		Update("revoked_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *ProfileRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	return repo.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now().UTC()).
		Delete(&RefreshTokenModel{}).Error
}
