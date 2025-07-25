package models

import "errors"

var (
	ErrEmptyPassword      = errors.New("password field is required")
	ErrEmptyName          = errors.New("name field is required")
	ErrInvalidCredentials = errors.New("credentials are invalid")
	ErrInvalidToken       = errors.New("token is invalid")
	ErrInvalidEmail       = errors.New("email is invalid")
	ErrExpToken           = errors.New("token expired")
	ErrNotUniqueEmail     = errors.New("email is not unique")
	ErrInvalidPassword    = errors.New("password must be in range of 8 and 72 bytes")
	ErrInvalidName        = errors.New("name must be in range of 4 and 72 bytes")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrUnexpected         = errors.New("internal server error: failed to handle incoming HTTP request")
	ErrTokenGenerateFail  = errors.New("failed to generate new token")
	ErrUserModelInvalid   = errors.New("user model is invalid")
	ErrCannotDeleteSelf   = errors.New("you cannot delete your own account while logged in as admin")
)
