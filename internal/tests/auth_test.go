package service

import (
	"authService/internal/domain"
	"authService/internal/repo"
	"authService/internal/service"
	"authService/internal/tests/mock"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

func TestLogin(t *testing.T) {
	testCases := []struct {
		name             string
		email            string
		password         string
		expectedHTTPcode int
		expectedErr      error
	}{
		{
			name:             "empty email",
			email:            "",
			password:         "password",
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrInvalidEmail,
		},
		{
			name:             "empty password",
			email:            "email@gmail.com",
			password:         "",
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrEmptyPassword,
		},
		{
			name:             "short password, less than 8",
			email:            "email@gmail.com",
			password:         "short",
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrInvalidPassword,
		},
		{
			name:             "long password, more than 72",
			email:            "email@gmail.com",
			password:         strings.Repeat("password", 10),
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrInvalidPassword,
		},
		{
			name:             "not exist email",
			email:            "uniqueMail@gmail.com",
			password:         "password",
			expectedHTTPcode: http.StatusUnauthorized,
			expectedErr:      repo.ErrUserNotExist,
		},
		{
			name:             "invalid password",
			email:            "defaultEmail@gmail.com",
			password:         "notvalidPassword",
			expectedHTTPcode: http.StatusUnauthorized,
			expectedErr:      service.ErrInvalidCredentials,
		}, {
			name:             "validLogin",
			email:            "defaultEmail@gmail.com",
			password:         "validPassword",
			expectedHTTPcode: http.StatusOK,
			expectedErr:      nil,
		},
	}

	authServ := service.NewAuthService(mock.NewMockUserRepo(), mock.NewMockTokenService(), slog.Default())
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, code, err := authServ.Login(tc.email, tc.password)
			if err != tc.expectedErr {
				t.Errorf("expected error = %v, got error = %v, err = %v", tc.expectedErr, err != nil, err)
			}
			if code != tc.expectedHTTPcode {
				t.Errorf("expected status code = %v, got code = %v, code = %v", tc.expectedHTTPcode, code != tc.expectedHTTPcode, code)
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
			name:             "empty email",
			userName:         "user name",
			email:            "",
			password:         "password",
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrInvalidEmail,
		}, {
			name:             "not unique email",
			userName:         "user name",
			email:            "ExistMail@gmail.com",
			password:         "password",
			expectedHTTPcode: http.StatusConflict,
			expectedErr:      service.ErrNotUniqueEmail,
		},
		{
			name:             "short password, less than 8",
			userName:         "user name",
			email:            "uniqueMail@gmail.com",
			password:         "short",
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrInvalidPassword,
		},
		{
			name:             "long password, more than 72",
			userName:         "user name",
			email:            "uniqueMail@gmail.com",
			password:         strings.Repeat("password", 10),
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrInvalidPassword,
		},
		{
			name:             "empty password",
			userName:         "user name",
			email:            "uniqueMail@gmail.com",
			password:         "",
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrEmptyPassword,
		},
		{
			name:             "empty name",
			userName:         "",
			password:         "password",
			email:            "uniqueMail@gmail.com",
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrEmptyName,
		},
		{
			name:             "short name",
			userName:         "123",
			password:         "password",
			email:            "uniqueMail@gmail.com",
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrInvalidName,
		},
		{
			name:             "very long name, more than 72",
			userName:         strings.Repeat("123456789", 9),
			password:         "password",
			email:            "uniqueMail@gmail.com",
			expectedHTTPcode: http.StatusBadRequest,
			expectedErr:      service.ErrInvalidName,
		}, {
			name:             "valid registration",
			userName:         "New User",
			email:            "uniqueMail@gmail.com",
			password:         "validPassword",
			expectedHTTPcode: http.StatusOK,
			expectedErr:      nil,
		},
	}
	authServ := service.NewAuthService(mock.NewMockUserRepo(), mock.NewMockTokenService(), slog.Default())
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, code, err := authServ.Register(tc.userName, tc.email, tc.password)
			if err != tc.expectedErr {
				t.Errorf("expected error = %v, got error = %v, err = %v", tc.expectedErr, err != nil, err)
			}
			if code != tc.expectedHTTPcode {
				t.Errorf("expected status code = %v, code = %v", tc.expectedHTTPcode, code)
			}
		})
	}
}

func TestCheckRole(t *testing.T) {
	tests := []struct {
		name             string
		token            string
		expectedUser     domain.User
		expectedHTTPcode int
		expectedErr      error
	}{
		{
			name:  "valid user token",
			token: "validToken",
			expectedUser: domain.User{
				ID:      1,
				Name:    "Test User",
				Email:   "beka123@gmail.com",
				IsAdmin: false,
			},
			expectedHTTPcode: http.StatusOK,
			expectedErr:      nil,
		},
		{
			name:             "invalid token",
			token:            "invalidToken",
			expectedUser:     domain.User{},
			expectedHTTPcode: http.StatusUnauthorized,
			expectedErr:      service.ErrInvalidToken,
		},
		{
			name:             "non-existent user",
			token:            "notExistToken",
			expectedUser:     domain.User{},
			expectedHTTPcode: http.StatusUnauthorized,
			expectedErr:      repo.ErrUserNotExist,
		},
		{
			name:  "admin user token",
			token: "adminToken",
			expectedUser: domain.User{
				ID:      1,
				Name:    "testName",
				Email:   "adminEmail@gmail.com",
				IsAdmin: true,
			},
			expectedHTTPcode: http.StatusOK,
			expectedErr:      nil,
		},
	}

	authServ := service.NewAuthService(mock.NewMockUserRepo(), mock.NewMockTokenService(), slog.Default())
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user, code, err := authServ.RoleCheck(tc.token)
			if err != tc.expectedErr {
				t.Errorf("expected error = %v, got error = %v, error = %v", tc.expectedErr, err != nil, err)
			}
			if code != tc.expectedHTTPcode {
				t.Errorf("expected status code = %d, got code = %v, code = %v", tc.expectedHTTPcode, tc.expectedHTTPcode != http.StatusOK, tc.expectedHTTPcode)
			}
			if err := EqualUsers(user, tc.expectedUser); err != nil {
				t.Error(err)
			}
		})
	}
}

func EqualUsers(got, expected domain.User) error {
	if got.ID != expected.ID {
		return fmt.Errorf("expected user ID = %d, got ID = %v", expected.ID, got.ID)
	}
	if got.IsAdmin != expected.IsAdmin {
		return fmt.Errorf("expected user isAdmin field = %t, got isAdmin field = %t", expected.IsAdmin, got.IsAdmin)
	}
	return nil
}
