package mock

import (
	"auth/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MockTokenService struct {
}

func NewMockTokenService() *MockTokenService {
	return &MockTokenService{}
}

func (s *MockTokenService) GenerateTokens(user models.User) (models.TokenPair, error) {
	return models.TokenPair{
		AccessToken:      "accessToken",
		AccessExpiresAt:  time.Now().Add(time.Minute * 15),
		RefreshToken:     "refreshToken",
		RefreshExpiresAt: time.Now().Add(time.Hour * 7),
	}, nil
}
func (s *MockTokenService) NewAccessClaim(user models.User) jwt.Claims {
	return jwt.MapClaims{}
}

func (s *MockTokenService) NewRefreshClaim(user models.User) jwt.Claims {
	return jwt.MapClaims{}
}

func (s *MockTokenService) Refresh(refreshToken string) (models.TokenPair, error) {
	return models.TokenPair{}, nil
}

func (s *MockTokenService) Validate(token string) (models.CustomClaims, error) {
	s.getSecret()
	isAdmin, isRefresh, email := false, false, "defaultEmail@gmail.com"
	switch token {
	case "adminToken":
		isAdmin = true
		email = "adminEmail@gmail.com"
	case "refresh":
		isRefresh = true
	case "invalidToken":
		return models.CustomClaims{}, models.ErrInvalidToken
	case "notExistToken":
		email = "uniqueMail@gmail.com"
	}

	return models.CustomClaims{
		Name:      "testName",
		Email:     email,
		IsAdmin:   isAdmin,
		IsRefresh: isRefresh,
	}, nil
}

func (s *MockTokenService) getSecret() string {
	return "secretKey"
}
