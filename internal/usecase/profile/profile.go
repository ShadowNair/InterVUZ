package profile

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/pbkdf2"
)

const (
	defaultAccessTokenTTL  = 15 * time.Minute
	defaultRefreshTokenTTL = 30 * 24 * time.Hour
	defaultPBKDF2Rounds    = 100_000
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrPasswordHashFailed = errors.New("password hash failed")
	ErrTokenGeneration    = errors.New("token generation failed")
)

type ProfileRepository interface {
	CreateProfile(ctx context.Context, profile domain.Profile) (*domain.Profile, error)
	GetProfileByEmail(ctx context.Context, email string) (*domain.Profile, error)
	GetProfileByID(ctx context.Context, id string) (*domain.Profile, error)
	CreateRefreshToken(ctx context.Context, token domain.RefreshToken) (*domain.RefreshToken, error)
	GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	DeleteExpiredRefreshTokens(ctx context.Context) error
}

type Config struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
}

type ProfileUCase struct {
	repo               ProfileRepository
	accessTokenSecret  []byte
	refreshTokenSecret []byte
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
	now                func() time.Time
}

type tokenClaims struct {
	TokenType string `json:"type"`
	jwt.RegisteredClaims
}

func New(repo ProfileRepository, cfg Config) *ProfileUCase {
	accessTTL := cfg.AccessTokenTTL
	if accessTTL <= 0 {
		accessTTL = defaultAccessTokenTTL
	}

	refreshTTL := cfg.RefreshTokenTTL
	if refreshTTL <= 0 {
		refreshTTL = defaultRefreshTokenTTL
	}

	accessSecret := cfg.AccessTokenSecret
	if accessSecret == "" {
		accessSecret = "dev-access-secret"
	}

	refreshSecret := cfg.RefreshTokenSecret
	if refreshSecret == "" {
		refreshSecret = "dev-refresh-secret"
	}

	return &ProfileUCase{
		repo:               repo,
		accessTokenSecret:  []byte(accessSecret),
		refreshTokenSecret: []byte(refreshSecret),
		accessTokenTTL:     accessTTL,
		refreshTokenTTL:    refreshTTL,
		now:                func() time.Time { return time.Now().UTC() },
	}
}

func (u *ProfileUCase) Register(ctx context.Context, request domain.RegisterRequest) (*domain.TokenPair, error) {
	request.Email = strings.TrimSpace(strings.ToLower(request.Email))
	request.Username = strings.TrimSpace(request.Username)
	request.Name = strings.TrimSpace(request.Name)
	request.Surname = strings.TrimSpace(request.Surname)
	request.Patronymic = strings.TrimSpace(request.Patronymic)

	if err := validateRegisterRequest(request); err != nil {
		return nil, err
	}

	existingProfile, err := u.repo.GetProfileByEmail(ctx, request.Email)
	switch {
	case err == nil && existingProfile != nil:
		return nil, ErrUserAlreadyExists
	case err != nil && !errors.Is(err, domain.ErrNotFound):
		return nil, err
	}

	salt, err := generateSalt()
	if err != nil {
		return nil, ErrPasswordHashFailed
	}

	passwordHash, err := hashPassword(request.Password, salt)
	if err != nil {
		return nil, ErrPasswordHashFailed
	}

	profile, err := u.repo.CreateProfile(ctx, domain.Profile{
		Email:        request.Email,
		Username:     request.Username,
		Name:         request.Name,
		Surname:      request.Surname,
		Patronymic:   request.Patronymic,
		Role:         "user",
		PasswordHash: passwordHash,
		PasswordSalt: salt,
	})
	if err != nil {
		if isDuplicateError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return u.issueTokenPair(ctx, profile.ID)
}

func (u *ProfileUCase) Login(ctx context.Context, request domain.LoginRequest) (*domain.TokenPair, error) {
	request.Email = strings.TrimSpace(strings.ToLower(request.Email))
	if request.Email == "" || request.Password == "" {
		return nil, ErrInvalidCredentials
	}

	profile, err := u.repo.GetProfileByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	passwordHash, err := hashPassword(request.Password, profile.PasswordSalt)
	if err != nil {
		return nil, ErrPasswordHashFailed
	}
	if passwordHash != profile.PasswordHash {
		return nil, ErrInvalidCredentials
	}

	return u.issueTokenPair(ctx, profile.ID)
}

func (u *ProfileUCase) Refresh(ctx context.Context, request domain.RefreshTokenRequest) (*domain.TokenPair, error) {
	rawToken := strings.TrimSpace(request.RefreshToken)
	if rawToken == "" {
		return nil, ErrInvalidToken
	}

	claims, err := u.parseToken(rawToken, "refresh", u.refreshTokenSecret)
	if err != nil {
		return nil, ErrInvalidToken
	}

	tokenHash := hashToken(rawToken)
	storedToken, err := u.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	now := u.now()
	if storedToken.RevokedAt != nil || !storedToken.ExpiresAt.After(now) || storedToken.ProfileID != claims.Subject {
		return nil, ErrInvalidToken
	}

	profile, err := u.repo.GetProfileByID(ctx, claims.Subject)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	if err := u.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, err
	}

	return u.issueTokenPair(ctx, profile.ID)
}

func (u *ProfileUCase) Logout(ctx context.Context, request domain.LogoutRequest) error {
	rawToken := strings.TrimSpace(request.RefreshToken)
	if rawToken == "" {
		return ErrInvalidToken
	}

	tokenHash := hashToken(rawToken)
	if err := u.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	return nil
}

func (u *ProfileUCase) Me(ctx context.Context, accessToken string) (*domain.ProfileResponse, error) {
	rawToken := strings.TrimSpace(accessToken)
	if rawToken == "" {
		return nil, ErrInvalidToken
	}

	claims, err := u.parseToken(rawToken, "access", u.accessTokenSecret)
	if err != nil {
		return nil, ErrInvalidToken
	}

	profile, err := u.repo.GetProfileByID(ctx, claims.Subject)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	profile.PasswordHash = ""
	profile.PasswordSalt = ""

	return &domain.ProfileResponse{Data: *profile}, nil
}

func (u *ProfileUCase) ToAccessTokenResponse(pair *domain.TokenPair) *domain.AccessTokenResponse {
	if pair == nil {
		return nil
	}

	return &domain.AccessTokenResponse{
		AccessToken:      pair.AccessToken,
		TokenType:        pair.TokenType,
		ExpiresIn:        pair.ExpiresIn,
		RefreshExpiresIn: pair.RefreshExpiresIn,
	}
}

func (u *ProfileUCase) issueTokenPair(ctx context.Context, profileID string) (*domain.TokenPair, error) {
	now := u.now()

	accessToken, err := u.generateToken(profileID, "access", u.accessTokenTTL, u.accessTokenSecret, now)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	refreshToken, err := u.generateToken(profileID, "refresh", u.refreshTokenTTL, u.refreshTokenSecret, now)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	if _, err := u.repo.CreateRefreshToken(ctx, domain.RefreshToken{
		ProfileID: profileID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: now.Add(u.refreshTokenTTL),
		CreatedAt: now,
	}); err != nil {
		return nil, err
	}

	_ = u.repo.DeleteExpiredRefreshTokens(ctx)

	return &domain.TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        int64(u.accessTokenTTL.Seconds()),
		RefreshExpiresIn: int64(u.refreshTokenTTL.Seconds()),
	}, nil
}

func (u *ProfileUCase) generateToken(profileID string, tokenType string, ttl time.Duration, secret []byte, now time.Time) (string, error) {
	claims := tokenClaims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   profileID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (u *ProfileUCase) parseToken(rawToken string, expectedType string, secret []byte) (*tokenClaims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &tokenClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid || claims.TokenType != expectedType || claims.Subject == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func validateRegisterRequest(request domain.RegisterRequest) error {
	if request.Email == "" || !strings.Contains(request.Email, "@") {
		return errors.New("email is required")
	}
	if len(request.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	return nil
}

func generateSalt() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashPassword(password string, salt string) (string, error) {
	if password == "" || salt == "" {
		return "", ErrPasswordHashFailed
	}

	derivedKey := pbkdf2.Key([]byte(password), []byte(salt), defaultPBKDF2Rounds, 32, sha256.New)
	return base64.RawURLEncoding.EncodeToString(derivedKey), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}

	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, "duplicate") || strings.Contains(lowered, "unique constraint")
}
