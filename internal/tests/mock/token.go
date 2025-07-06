package mock

import (
	"authService/internal/domain"
	"authService/internal/service"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MockTokenService struct {
}

func NewMockTokenService() *MockTokenService {
	return &MockTokenService{}
}

func (s *MockTokenService) GenerateTokens(user domain.User) (domain.TokenPair, error) {
	return domain.TokenPair{
		AccessToken:      "accessToken",
		AccessExpiresAt:  time.Now().Add(time.Minute * 15),
		RefreshToken:     "refreshToken",
		RefreshExpiresAt: time.Now().Add(time.Hour * 7),
	}, nil
}
func (s *MockTokenService) NewAccessClaim(user domain.User) jwt.Claims {
	return jwt.MapClaims{}
}

func (s *MockTokenService) NewRefreshClaim(user domain.User) jwt.Claims {
	return jwt.MapClaims{}
}
func (s *MockTokenService) Refresh(refreshToken string) (domain.TokenPair, int, error) {
	return domain.TokenPair{}, http.StatusOK, nil
}
func (s *MockTokenService) Validate(token string) (domain.CustomClaims, error) {
	s.getSecret()
	isAdmin, isRefresh, email := false, false, "defaultEmail@gmail.com"
	switch token {
	case "adminToken":
		isAdmin = true
		email = "adminEmail@gmail.com"
	case "refresh":
		isRefresh = true
	case "invalidToken":
		return domain.CustomClaims{}, service.ErrInvalidToken
	case "notExistToken":
		email = "uniqueMail@gmail.com"
	}

	return domain.CustomClaims{
		Name:      "testName",
		Email:     email,
		IsAdmin:   isAdmin,
		IsRefresh: isRefresh,
	}, nil
}

func (s *MockTokenService) getSecret() string {
	return "secretKey"
}
