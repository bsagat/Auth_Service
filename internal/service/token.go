package service

import (
	"authService/internal/dal"
	"authService/internal/domain"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	UserDal    *dal.UserDal
	RefreshTTL time.Duration
	AccessTTL  time.Duration
	log        *slog.Logger
	secret     string
}

func NewTokenService(secret string, UserDal *dal.UserDal, RefreshTTL time.Duration, AccessTTL time.Duration, log *slog.Logger) *TokenService {
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
	var signed []string
	for _, claim := range []jwt.Claims{s.NewAccessClaim(user), s.NewRefreshClaim(user)} {
		// Подпись каждого jwt токена
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
		signedToken, err := token.SignedString([]byte(s.getSecret()))
		if err != nil {
			s.log.Error("Failed to sign string", "error", err)
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

func (s *TokenService) NewAccessClaim(user domain.User) jwt.Claims {
	return jwt.MapClaims{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"is_admin":   user.IsAdmin,
		"is_refresh": false,
		"exp":        time.Now().Add(s.AccessTTL).Unix(),
	}
}

func (s *TokenService) NewRefreshClaim(user domain.User) jwt.Claims {
	return jwt.MapClaims{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"is_admin":   user.IsAdmin,
		"is_refresh": true,
		"exp":        time.Now().Add(s.RefreshTTL).Unix(),
	}
}

func (s *TokenService) Refresh(refreshToken string) (domain.TokenPair, int, error) {
	const op = "AuthService.RefreshToken"
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
		if errors.Is(err, dal.ErrUserNotExist) {
			log.Error("User is not exist")
			return domain.TokenPair{}, http.StatusBadRequest, dal.ErrUserNotExist
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
		return domain.CustomClaims{}, err
	}

	if !parsedToken.Valid {
		return domain.CustomClaims{}, errors.New("token is invalid")
	}

	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return domain.CustomClaims{}, errors.New("invalid token claims structure")
	}

	var claims domain.CustomClaims

	// Извлекаем Email
	if email, ok := mapClaims["email"].(string); ok {
		claims.Email = email
	} else {
		return domain.CustomClaims{}, errors.New("invalid or missing 'email' in token claims")
	}

	// Извлекаем Name
	if name, ok := mapClaims["name"].(string); ok {
		claims.Name = name
	}

	// Извлекаем IsAdmin
	if isAdmin, ok := mapClaims["is_admin"].(bool); ok {
		claims.IsAdmin = isAdmin
	}

	// Извлекаем exp и проверяем время
	expFloat, ok := mapClaims["exp"].(float64)
	if !ok {
		return domain.CustomClaims{}, errors.New("invalid or missing 'exp' in token claims")
	}
	expTime := time.Unix(int64(expFloat), 0)
	claims.ExpiresAt = jwt.NewNumericDate(expTime)

	if time.Now().After(expTime) {
		return domain.CustomClaims{}, errors.New("token expired")
	}

	return claims, nil
}
