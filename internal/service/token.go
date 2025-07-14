package service

import (
	"auth/internal/adapters/repo"
	"auth/internal/domain"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	UserDal    domain.UserRepo
	RefreshTTL time.Duration
	AccessTTL  time.Duration
	log        *slog.Logger
	secret     string
}

func NewTokenService(secret string, UserDal domain.UserRepo, RefreshTTL time.Duration, AccessTTL time.Duration, log *slog.Logger) *TokenService {
	return &TokenService{
		UserDal:    UserDal,
		RefreshTTL: RefreshTTL,
		AccessTTL:  AccessTTL,
		secret:     secret,
		log:        log,
	}
}

func (s *TokenService) getSecret() string {
	return s.secret
}

func (s *TokenService) GenerateTokens(user domain.User) (domain.TokenPair, error) {
	const op = "TokenService.GenerateTokens"
	log := s.log.With(
		slog.String("op", op),
	)

	var signed []string
	for _, claim := range []jwt.Claims{NewAccessClaim(user, s.AccessTTL), NewRefreshClaim(user, s.RefreshTTL)} {
		// Подпись каждого jwt токена
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
		signedToken, err := token.SignedString([]byte(s.getSecret()))
		if err != nil {
			log.Error("Failed to sign string", "error", err)
			return domain.TokenPair{}, err
		}
		signed = append(signed, signedToken)
	}
	return domain.TokenPair{
		AccessExpiresAt:  time.Now().Add(s.AccessTTL),
		RefreshExpiresAt: time.Now().Add(s.RefreshTTL),
		AccessToken:      signed[0],
		RefreshToken:     signed[1],
	}, nil
}

func NewAccessClaim(user domain.User, accessTTL time.Duration) jwt.Claims {
	return jwt.MapClaims{
		"ID":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"is_admin":   user.IsAdmin,
		"is_refresh": false,
		"exp":        time.Now().Add(accessTTL).Unix(),
	}
}

func NewRefreshClaim(user domain.User, refreshTTL time.Duration) jwt.Claims {
	return jwt.MapClaims{
		"ID":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"is_admin":   user.IsAdmin,
		"is_refresh": true,
		"exp":        time.Now().Add(refreshTTL).Unix(),
	}
}

func (s *TokenService) Refresh(refreshToken string) (domain.TokenPair, int, error) {
	const op = "TokenService.RefreshToken"
	log := s.log.With(
		slog.String("op", op),
	)
	log.Info("Token refresh started")

	claims, err := s.Validate(refreshToken)
	if err != nil {
		log.Error("Refresh token is invalid", "error", err)
		return domain.TokenPair{}, http.StatusUnauthorized, err
	}

	// Проверяем существует ли пользователь
	user, err := s.UserDal.GetUser(claims.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return domain.TokenPair{}, http.StatusUnauthorized, repo.ErrUserNotExist
		}
		log.Error("Failed to check user uniqueness", "error", err)
		return domain.TokenPair{}, http.StatusInternalServerError, errors.New("failed to check user uniqueness")
	}

	pair, err := s.GenerateTokens(user)
	if err != nil {
		log.Error("Failed to generate tokens", "error", err)
		return domain.TokenPair{}, http.StatusInternalServerError, err
	}

	return pair, http.StatusOK, nil
}

func (s *TokenService) Validate(token string) (domain.CustomClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.getSecret()), nil
	})
	if err != nil {
		s.log.Error("Failed to parse with claims", "error", err)
		return domain.CustomClaims{}, ErrInvalidToken
	}

	if !parsedToken.Valid {
		return domain.CustomClaims{}, ErrInvalidToken
	}

	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return domain.CustomClaims{}, errors.New("invalid token claims structure")
	}

	var claims domain.CustomClaims
	var invOrMissingForm string = "invalid or missing '%s' in token claims"

	// Извлекаем Email
	if email, ok := mapClaims["email"].(string); ok {
		claims.Email = email
	} else {

		return domain.CustomClaims{}, fmt.Errorf(invOrMissingForm, "email")
	}

	// Извлекаем Name
	if name, ok := mapClaims["name"].(string); ok {
		claims.Name = name
	} else {
		return domain.CustomClaims{}, fmt.Errorf(invOrMissingForm, "name")
	}

	// Извлекаем IsAdmin
	if isAdmin, ok := mapClaims["is_admin"].(bool); ok {
		claims.IsAdmin = isAdmin
	} else {
		return domain.CustomClaims{}, fmt.Errorf(invOrMissingForm, "is_admin")
	}

	if isRefresh, ok := mapClaims["is_refresh"].(bool); ok {
		claims.IsRefresh = isRefresh
	} else {
		return domain.CustomClaims{}, fmt.Errorf(invOrMissingForm, "is_refresh")
	}

	// Извлекаем exp и проверяем время
	expFloat, ok := mapClaims["exp"].(float64)
	if !ok {
		return domain.CustomClaims{}, fmt.Errorf(invOrMissingForm, "exp")
	}
	expTime := time.Unix(int64(expFloat), 0)
	claims.ExpiresAt = jwt.NewNumericDate(expTime)

	if time.Now().After(expTime) {
		return domain.CustomClaims{}, errors.New("token expired")
	}

	return claims, nil
}
