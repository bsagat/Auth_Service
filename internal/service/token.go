package service

import (
	"auth/internal/adapters/repo"
	"auth/internal/domain/models"
	"auth/internal/domain/ports"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	UserDal    ports.UserRepo
	RefreshTTL time.Duration
	AccessTTL  time.Duration
	log        *slog.Logger
	secret     string
}

func NewTokenService(secret string, UserDal ports.UserRepo, RefreshTTL time.Duration, AccessTTL time.Duration, log *slog.Logger) *TokenService {
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

func (s *TokenService) GenerateTokens(user models.User) (models.TokenPair, error) {
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
			return models.TokenPair{}, err
		}
		signed = append(signed, signedToken)
	}
	return models.TokenPair{
		AccessExpiresAt:  time.Now().Add(s.AccessTTL),
		RefreshExpiresAt: time.Now().Add(s.RefreshTTL),
		AccessToken:      signed[0],
		RefreshToken:     signed[1],
	}, nil
}

func NewAccessClaim(user models.User, accessTTL time.Duration) jwt.Claims {
	return jwt.MapClaims{
		"ID":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"is_admin":   user.IsAdmin,
		"is_refresh": false,
		"role":       user.Role,
		"exp":        time.Now().Add(accessTTL).Unix(),
	}
}

func NewRefreshClaim(user models.User, refreshTTL time.Duration) jwt.Claims {
	return jwt.MapClaims{
		"ID":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"is_admin":   user.IsAdmin,
		"is_refresh": true,
		"role":       user.Role,
		"exp":        time.Now().Add(refreshTTL).Unix(),
	}
}

func (s *TokenService) Refresh(refreshToken string) (models.TokenPair, error) {
	const op = "TokenService.RefreshToken"
	log := s.log.With(
		slog.String("op", op),
	)
	log.Info("Token refresh started")

	claims, err := s.Validate(refreshToken)
	if err != nil {
		log.Error("Refresh token is invalid", "error", err)
		return models.TokenPair{}, models.ErrInvalidToken
	}

	// Проверяем существует ли пользователь
	user, err := s.UserDal.GetUser(claims.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return models.TokenPair{}, repo.ErrUserNotExist
		}
		log.Error("Failed to check user uniqueness", "error", err)
		return models.TokenPair{}, models.ErrUnexpected
	}

	pair, err := s.GenerateTokens(user)
	if err != nil {
		log.Error("Failed to generate tokens", "error", err)
		return models.TokenPair{}, models.ErrUnexpected
	}

	return pair, nil
}

func (s *TokenService) Validate(token string) (models.CustomClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.getSecret()), nil
	})
	if err != nil {
		s.log.Error("Failed to parse with claims", "error", err)
		return models.CustomClaims{}, models.ErrInvalidToken
	}

	if !parsedToken.Valid {
		return models.CustomClaims{}, models.ErrInvalidToken
	}

	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return models.CustomClaims{}, models.ErrInvalidToken
	}

	var claims models.CustomClaims
	var invOrMissingForm string = "invalid or missing '%s' in token claims"

	// Извлекаем Name
	if Id, ok := mapClaims["ID"].(float64); ok {
		claims.ID = int(Id)
	} else {
		return models.CustomClaims{}, fmt.Errorf(invOrMissingForm, "ID")
	}

	if role, ok := mapClaims["role"].(string); ok {
		claims.Role = role
	} else {
		return models.CustomClaims{}, fmt.Errorf(invOrMissingForm, "role")
	}

	// Извлекаем Email
	if email, ok := mapClaims["email"].(string); ok {
		claims.Email = email
	} else {
		return models.CustomClaims{}, fmt.Errorf(invOrMissingForm, "email")
	}

	// Извлекаем Name
	if name, ok := mapClaims["name"].(string); ok {
		claims.Name = name
	} else {
		return models.CustomClaims{}, fmt.Errorf(invOrMissingForm, "name")
	}

	// Извлекаем IsAdmin
	if isAdmin, ok := mapClaims["is_admin"].(bool); ok {
		claims.IsAdmin = isAdmin
	} else {
		return models.CustomClaims{}, fmt.Errorf(invOrMissingForm, "is_admin")
	}

	if isRefresh, ok := mapClaims["is_refresh"].(bool); ok {
		claims.IsRefresh = isRefresh
	} else {
		return models.CustomClaims{}, fmt.Errorf(invOrMissingForm, "is_refresh")
	}

	// Извлекаем exp и проверяем время
	expFloat, ok := mapClaims["exp"].(float64)
	if !ok {
		return models.CustomClaims{}, fmt.Errorf(invOrMissingForm, "exp")
	}
	expTime := time.Unix(int64(expFloat), 0)
	claims.ExpiresAt = jwt.NewNumericDate(expTime)

	if time.Now().After(expTime) {
		return models.CustomClaims{}, models.ErrExpToken
	}

	return claims, nil
}
