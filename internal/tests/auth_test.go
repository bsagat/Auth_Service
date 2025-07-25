package service

import (
	"auth/internal/adapters/repo"
	"auth/internal/adapters/transport/http/routers"
	"auth/internal/domain/models"
	"auth/internal/service"
	"auth/internal/tests/mock"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

func TestValidateCredentials(t *testing.T) {
	testCases := []struct {
		name             string
		userName         string
		email            string
		password         string
		expectedHTTPcode int
		expectedErr      error
	}{
		{
			name:        "empty email",
			userName:    "New User",
			email:       "",
			password:    "password",
			expectedErr: models.ErrInvalidEmail,
		},
		{
			name:        "empty password",
			userName:    "New User",
			email:       "email@gmail.com",
			password:    "",
			expectedErr: models.ErrEmptyPassword,
		},
		{
			name:        "short password, less than 8",
			userName:    "New User",
			email:       "email@gmail.com",
			password:    "short",
			expectedErr: models.ErrInvalidPassword,
		},
		{
			name:        "long password, more than 72",
			userName:    "New User",
			email:       "email@gmail.com",
			password:    strings.Repeat("password", 10),
			expectedErr: models.ErrInvalidPassword,
		},
		{
			name:        "validLogin",
			userName:    "New User",
			email:       "defaultEmail@gmail.com",
			password:    "validPassword",
			expectedErr: nil,
		},
		{
			name:        "empty name",
			userName:    "",
			password:    "password",
			email:       "uniqueMail@gmail.com",
			expectedErr: models.ErrEmptyName,
		},
		{
			name:        "short name",
			userName:    "123",
			password:    "password",
			email:       "uniqueMail@gmail.com",
			expectedErr: models.ErrInvalidName,
		},
		{
			name:        "very long name, more than 72",
			userName:    strings.Repeat("123456789", 9),
			password:    "password",
			email:       "uniqueMail@gmail.com",
			expectedErr: models.ErrInvalidName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := routers.ValidateCredentials(tc.userName, tc.email, tc.password)
			if err != tc.expectedErr {
				t.Errorf("expected error = %v, got error = %v, err = %v", tc.expectedErr, err != nil, err)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	testCases := []struct {
		name             string
		email            string
		password         string
		expectedHTTPcode int
		expectedErr      error
	}{
		{
			name:        "not exist email",
			email:       "uniqueMail@gmail.com",
			password:    "password",
			expectedErr: repo.ErrUserNotExist,
		},
		{
			name:        "invalid password",
			email:       "defaultEmail@gmail.com",
			password:    "notvalidPassword",
			expectedErr: models.ErrInvalidCredentials,
		}, {
			name:        "validLogin",
			email:       "defaultEmail@gmail.com",
			password:    "validPassword",
			expectedErr: nil,
		},
	}
	authServ := service.NewAuthService(mock.NewMockUserRepo(), mock.NewMockTokenService(), slog.Default())
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := authServ.Login(tc.email, tc.password)
			if err != tc.expectedErr {
				t.Errorf("expected error = %v, got error = %v, err = %v", tc.expectedErr, err != nil, err)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name             string
		userName         string
		email            string
		password         string
		expectedHTTPcode int
		expectedErr      error
	}{
		{
			name:        "not unique email",
			userName:    "user name",
			email:       "ExistMail@gmail.com",
			password:    "password",
			expectedErr: models.ErrNotUniqueEmail,
		},
		{
			name:        "valid registration",
			userName:    "New User",
			email:       "uniqueMail@gmail.com",
			password:    "validPassword",
			expectedErr: nil,
		},
	}
	authServ := service.NewAuthService(mock.NewMockUserRepo(), mock.NewMockTokenService(), slog.Default())
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := authServ.Register(tc.userName, tc.email, tc.password)
			if err != tc.expectedErr {
				t.Errorf("expected error = %v, got error = %v, err = %v", tc.expectedErr, err != nil, err)
			}
		})
	}
}

func TestCheckRole(t *testing.T) {
	tests := []struct {
		name             string
		token            string
		expectedUser     models.User
		expectedHTTPcode int
		expectedErr      error
	}{
		{
			name:  "valid user token",
			token: "validToken",
			expectedUser: models.User{
				ID:      1,
				Name:    "Test User",
				Email:   "beka123@gmail.com",
				IsAdmin: false,
			},
			expectedHTTPcode: http.StatusOK,
			expectedErr:      nil,
		},
		{
			name:         "invalid token",
			token:        "invalidToken",
			expectedUser: models.User{},
			expectedErr:  models.ErrInvalidToken,
		},
		{
			name:         "non-existent user",
			token:        "notExistToken",
			expectedUser: models.User{},
			expectedErr:  repo.ErrUserNotExist,
		},
		{
			name:  "admin user token",
			token: "adminToken",
			expectedUser: models.User{
				ID:      1,
				Name:    "testName",
				Email:   "adminEmail@gmail.com",
				IsAdmin: true,
			},
			expectedErr: nil,
		},
	}

	authServ := service.NewAuthService(mock.NewMockUserRepo(), mock.NewMockTokenService(), slog.Default())
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user, err := authServ.RoleCheck(tc.token)
			if err != tc.expectedErr {
				t.Errorf("expected error = %v, got error = %v, error = %v", tc.expectedErr, err != nil, err)
			}
			if err := EqualUsers(user, tc.expectedUser); err != nil {
				t.Error(err)
			}
		})
	}
}

func EqualUsers(got, expected models.User) error {
	if got.ID != expected.ID {
		return fmt.Errorf("expected user ID = %d, got ID = %v", expected.ID, got.ID)
	}
	if got.IsAdmin != expected.IsAdmin {
		return fmt.Errorf("expected user isAdmin field = %t, got isAdmin field = %t", expected.IsAdmin, got.IsAdmin)
	}
	return nil
}
