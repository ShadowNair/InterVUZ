package common

import (
	"errors"
	"net/http"
	"strings"

	httputil "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/common"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/profile"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
)

func WriteUseCaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, profile.ErrUserAlreadyExists):
		httputil.WriteError(w, http.StatusConflict, "conflict", err.Error())
	case errors.Is(err, profile.ErrInvalidCredentials),
		errors.Is(err, profile.ErrInvalidToken),
		errors.Is(err, profile.ErrUserNotFound):
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
	case strings.Contains(strings.ToLower(err.Error()), "required"),
		strings.Contains(strings.ToLower(err.Error()), "password"),
		strings.Contains(strings.ToLower(err.Error()), "email"):
		httputil.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
	default:
		httputil.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func ExtractBearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", profile.ErrInvalidToken
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", profile.ErrInvalidToken
	}

	return strings.TrimSpace(parts[1]), nil
}

func SetRefreshTokenCookie(w http.ResponseWriter, pair *domain.TokenPair) {
	if pair == nil || pair.RefreshToken == "" {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    pair.RefreshToken,
		Path:     "/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   int(pair.RefreshExpiresIn),
	})
}

func SetAccessTokenCookie(w http.ResponseWriter, pair *domain.TokenPair) {
	if pair == nil || pair.AccessToken == "" {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    pair.AccessToken,
		Path:     "/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   int(pair.ExpiresIn),
	})
}

func ClearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   -1,
	})
}

func ClearAccessTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   -1,
	})
}

func ExtractRefreshTokenCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshTokenCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", profile.ErrInvalidToken
	}

	return strings.TrimSpace(cookie.Value), nil
}

func ExtractAccessToken(r *http.Request) (string, error) {
	if cookie, err := r.Cookie(AccessTokenCookieName); err == nil && strings.TrimSpace(cookie.Value) != "" {
		return strings.TrimSpace(cookie.Value), nil
	}

	return ExtractBearerToken(r.Header.Get("Authorization"))
}
