package profile_repo

import (
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type ProfileModel struct {
	ID           string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	PasswordSalt string    `gorm:"column:password_salt"`
	Username     *string   `gorm:"column:username"`
	Name         *string   `gorm:"column:name"`
	Surname      *string   `gorm:"column:surname"`
	Patronymic   *string   `gorm:"column:patronymic"`
	AvatarURL    *string   `gorm:"column:avatar_url"`
	Role         string    `gorm:"column:role"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

type RefreshTokenModel struct {
	ID        string     `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	ProfileID string     `gorm:"column:profile_id;type:uuid"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
}

func (ProfileModel) TableName() string {
	return "profiles"
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

func ToProfileModel(profile domain.Profile) ProfileModel {
	return ProfileModel{
		ID:           profile.ID,
		Email:        profile.Email,
		PasswordHash: profile.PasswordHash,
		PasswordSalt: profile.PasswordSalt,
		Username:     nullableString(profile.Username),
		Name:         nullableString(profile.Name),
		Surname:      nullableString(profile.Surname),
		Patronymic:   nullableString(profile.Patronymic),
		AvatarURL:    nullableString(profile.AvatarURL),
		Role:         profile.Role,
		CreatedAt:    profile.CreatedAt,
		UpdatedAt:    profile.UpdatedAt,
	}
}

func ToDomainProfile(model ProfileModel) domain.Profile {
	return domain.Profile{
		ID:           model.ID,
		Email:        model.Email,
		Username:     derefString(model.Username),
		Name:         derefString(model.Name),
		Surname:      derefString(model.Surname),
		Patronymic:   derefString(model.Patronymic),
		AvatarURL:    derefString(model.AvatarURL),
		Role:         model.Role,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		PasswordHash: model.PasswordHash,
		PasswordSalt: model.PasswordSalt,
	}
}

func ToRefreshTokenModel(token domain.RefreshToken) RefreshTokenModel {
	return RefreshTokenModel{
		ID:        token.ID,
		ProfileID: token.ProfileID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
		RevokedAt: token.RevokedAt,
	}
}

func ToDomainRefreshToken(model RefreshTokenModel) domain.RefreshToken {
	return domain.RefreshToken{
		ID:        model.ID,
		ProfileID: model.ProfileID,
		TokenHash: model.TokenHash,
		ExpiresAt: model.ExpiresAt,
		CreatedAt: model.CreatedAt,
		RevokedAt: model.RevokedAt,
	}
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
