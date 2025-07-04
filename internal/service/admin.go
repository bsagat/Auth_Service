package service

import (
	"authService/internal/domain"
	"authService/internal/repo"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type AdminService struct {
	UserDal   *repo.UserDal
	TokenServ *TokenService
	log       *slog.Logger
}

func NewAdminService(UserDal *repo.UserDal, TokenServ *TokenService, log *slog.Logger) *AdminService {
	return &AdminService{
		UserDal:   UserDal,
		TokenServ: TokenServ,
		log:       log,
	}
}

func (s *AdminService) GetUser(userID int, access string) (domain.User, int, error) {
	const op = "AdminService.GetUser"
	log := s.log.With(
		slog.String("op", op),
		slog.Int("ID", userID),
	)

	// Валидируем токен
	claims, err := s.TokenServ.Validate(access)
	if err != nil {
		log.Error("Access token is invalid", "error", err)
		return domain.User{}, http.StatusUnauthorized, err
	}

	// Проверяем права пользователя
	if !claims.IsAdmin {
		log.Error("User is not administrator")
		return domain.User{}, http.StatusForbidden, errors.New("permission denied")
	}

	// Получаем user-а
	existUser, err := s.UserDal.GetUser(claims.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return domain.User{}, http.StatusNotFound, repo.ErrUserNotExist
		}
		log.Error("Failed to check user uniqueness", "error", err)
		return domain.User{}, http.StatusInternalServerError, errors.New("failed to check user uniqueness")
	}

	return existUser, http.StatusOK, nil
}

func (s *AdminService) DeleteUser(userID int, access string) (int, error) {
	const op = "AdminService.DeleteUser"
	log := s.log.With(
		slog.String("op", op),
		slog.Int("ID", userID),
	)

	// Валидируем токен
	claims, err := s.TokenServ.Validate(access)
	if err != nil {
		log.Error("Access token is invalid", "error", err)
		return http.StatusUnauthorized, err
	}

	// Проверяем права пользователя
	if !claims.IsAdmin {
		log.Error("User is not administrator")
		return http.StatusForbidden, errors.New("permission denied")
	}

	// Удаляем пользователя
	if err := s.UserDal.DeleteUser(userID); err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return http.StatusNotFound, err
		}
		log.Error("Failed to delete user data", "error", err)
		return http.StatusInternalServerError, err
	}
	return http.StatusNoContent, nil
}

func (s *AdminService) UpdateUser(user domain.User, access string) (int, error) {
	const op = "AdminService.UpdateUser"
	log := s.log.With(
		slog.String("op", op),
		slog.Int("ID", user.ID),
		slog.String("name", user.Name),
	)

	// Валидируем запрос
	if err := ValidateUpdateUser(user); err != nil {
		log.Error("Update request is invalid", "error", err)
		return http.StatusBadRequest, err
	}

	// Валидируем токен
	claims, err := s.TokenServ.Validate(access)
	if err != nil {
		log.Error("Access token is invalid", "error", err)
		return http.StatusUnauthorized, err
	}

	// Проверяем права пользователя
	if !claims.IsAdmin {
		log.Error("User is not administrator")
		return http.StatusForbidden, errors.New("permission denied")
	}

	// Пока что обновляем name
	// Можно полностью, когда будет доступен tokens-black-list
	if err := s.UserDal.UpdateUserName(user.Name, user.ID); err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return http.StatusNotFound, err
		}
		log.Error("Failed to delete user data", "error", err)
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func ValidateUpdateUser(user domain.User) error {
	if user.ID == 0 {
		return fmt.Errorf("user id field is reqired")
	}
	if len(user.Name) == 0 {
		return errors.New("user id field is reqired")
	}
	return nil
}
