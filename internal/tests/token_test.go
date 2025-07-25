package service

import (
	"auth/internal/domain/models"
	"auth/internal/service"
	"auth/internal/tests/mock"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestGenerateAndValidateTokens(t *testing.T) {
	user := models.User{
		ID:      1,
		Name:    "Test User",
		Email:   "test@example.com",
		IsAdmin: false,
	}

	tokenService := service.NewTokenService(
		"supersecretkey",
		nil, // UserRepo не нужен для GenerateTokens и Validate
		time.Minute*5,
		time.Minute*5,
		slog.Default(),
	)

	tokens, err := tokenService.GenerateTokens(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens, got %+v", tokens)
	}

	claims, err := tokenService.Validate(tokens.AccessToken)
	if err != nil {
		t.Fatalf("expected no error during validation, got %v", err)
	}

	if claims.Email != user.Email {
		t.Errorf("expected email %v, got %v", user.Email, claims.Email)
	}
	if claims.Name != user.Name {
		t.Errorf("expected name %v, got %v", user.Name, claims.Name)
	}
	if claims.IsAdmin != user.IsAdmin {
		t.Errorf("expected IsAdmin %v, got %v", user.IsAdmin, claims.IsAdmin)
	}
	if claims.IsRefresh != false {
		t.Errorf("expected IsRefresh false, got %v", claims.IsRefresh)
	}
}

func TestValidate_InvalidToken(t *testing.T) {
	tokenService := service.NewTokenService(
		"supersecretkey",
		nil,
		time.Minute,
		time.Minute,
		slog.Default(),
	)

	_, err := tokenService.Validate("invalid.super.token")
	if err == nil || !strings.Contains(err.Error(), "token is invalid") {
		t.Errorf("expected invalid token error, got %v", err)
	}
}

func TestRefresh_Success(t *testing.T) {
	mockDal := mock.NewMockUserRepo()
	tokenService := service.NewTokenService(
		"supersecretkey",
		mockDal,
		time.Minute*5,
		time.Minute*5,
		slog.Default(),
	)

	user := models.User{
		ID:      1,
		Name:    "Test User",
		Email:   "test@example.com",
		IsAdmin: false,
	}

	tokens, err := tokenService.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	refreshed, err := tokenService.Refresh(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Errorf("expected non-empty tokens, got %+v", refreshed)
	}
}

func TestRefresh_UserNotExist(t *testing.T) {
	mockDal := mock.NewMockUserRepo()
	tokenService := service.NewTokenService(
		"supersecretkey",
		mockDal,
		time.Minute*5,
		time.Minute*5,
		slog.Default(),
	)

	user := models.User{
		ID:      1,
		Name:    "Test User",
		Email:   "uniqueMail@gmail.com",
		IsAdmin: false,
	}

	tokens, err := tokenService.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}

	_, err = tokenService.Refresh(tokens.RefreshToken)
	if err == nil {
		t.Fatal("expected error for non-existing user, got nil")
	}
}
