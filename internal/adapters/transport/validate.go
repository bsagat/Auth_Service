package validate

import (
	"auth/internal/domain/models"
	"errors"
	"fmt"
	"net/mail"
	"slices"
)

func UserReq(userID int, name, role string) error {
	if userID == 0 {
		return errors.New("user id field is reqired")
	}
	if len(name) == 0 {
		return errors.New("user name field is reqired")
	}

	if len(name) < 4 || len(name) > 72 {
		return models.ErrInvalidName
	}

	return Role(role)
}

func Credentials(name, email, password, role string) error {
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

	if err := Role(role); err != nil {
		return err
	}

	return nil
}

func Role(role string) error {
	if len(role) == 0 {
		return errors.New("user role field is reqired")
	}

	if slices.Contains([]string{models.AdminRole, models.UserRole}, role) {
		return nil
	}
	return fmt.Errorf("%w: %s", models.ErrInvalidRole, role)
}
