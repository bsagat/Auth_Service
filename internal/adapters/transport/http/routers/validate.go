package routers

import (
	"auth/internal/adapters/transport/http/dto"
	"auth/internal/domain/models"
	"errors"
	"fmt"
	"net/mail"
	"slices"
)

func ValidateUserReq(user dto.UpdateUserReq) error {
	if user.ID == 0 {
		return errors.New("user id field is reqired")
	}
	if len(user.Name) == 0 {
		return errors.New("user name field is reqired")
	}

	if len(user.Name) < 4 || len(user.Name) > 72 {
		return models.ErrInvalidName
	}

	if len(user.Role) == 0 {
		return errors.New("user role field is reqired")
	}

	return ValidateRole(user.Role)
}

func ValidateCredentials(name, email, password, role string) error {
	// Name check
	if len(name) == 0 {
		return models.ErrEmptyName
	}

	if len(name) < 4 || len(name) > 72 {
		return models.ErrInvalidName
	}

	// Email check
	_, err := mail.ParseAddress(email)
	if err != nil || len(email) > 255 {
		return models.ErrInvalidEmail
	}

	// Password check
	if len(password) == 0 {
		return models.ErrEmptyPassword
	}

	if len(password) > 72 || len(password) < 8 {
		return models.ErrInvalidPassword
	}

	if err := ValidateRole(role); err != nil {
		return err
	}

	return nil
}

func ValidateRole(role string) error {
	if slices.Contains([]string{models.AdminRole, models.UserRole}, role) {
		return nil
	}
	return fmt.Errorf("%w: %s", models.ErrInvalidRole, role)
}
