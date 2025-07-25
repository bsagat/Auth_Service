package service

import (
	"auth/internal/adapters/repo"
	"auth/internal/domain/models"
	"errors"
	"log/slog"
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

func (s *AdminService) GetUser(userID int, access string) (models.User, error) {
	const op = "AdminService.GetUser"
	log := s.log.With(
		slog.String("op", op),
		slog.Int("ID", userID),
	)

	// Валидируем токен
	claims, err := s.TokenServ.Validate(access)
	if err != nil {
		log.Error("Access token is invalid", "error", err)
		return models.User{}, models.ErrInvalidToken
	}

	// Проверяем права пользователя
	if !claims.IsAdmin {
		log.Error("User is not administrator")
		return models.User{}, models.ErrPermissionDenied
	}

	// Получаем user-а
	existUser, err := s.UserDal.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return models.User{}, repo.ErrUserNotExist
		}
		log.Error("Failed to check user uniqueness", "error", err)
		return models.User{}, models.ErrUnexpected
	}

	return existUser, nil
}

func (s *AdminService) DeleteUser(userID int, access string) error {
	const op = "AdminService.DeleteUser"
	log := s.log.With(
		slog.String("op", op),
		slog.Int("ID", userID),
	)

	// Валидируем токен
	claims, err := s.TokenServ.Validate(access)
	if err != nil {
		log.Error("Access token is invalid", "error", err)
		return models.ErrInvalidToken
	}

	// Проверяем права пользователя
	if !claims.IsAdmin {
		log.Error("User is not administrator")
		return models.ErrPermissionDenied
	}

	if claims.ID == userID {
		return models.ErrCannotDeleteSelf
	}

	// Удаляем пользователя
	if err := s.UserDal.DeleteUser(userID); err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return repo.ErrUserNotExist
		}
		log.Error("Failed to delete user data", "error", err)
		return models.ErrUnexpected
	}
	return nil
}

func (s *AdminService) UpdateUser(user models.User, access string) error {
	const op = "AdminService.UpdateUser"
	log := s.log.With(
		slog.String("op", op),
		slog.Int("ID", user.ID),
		slog.String("name", user.Name),
	)

	// Валидируем токен
	claims, err := s.TokenServ.Validate(access)
	if err != nil {
		log.Error("Access token is invalid", "error", err)
		return models.ErrInvalidToken
	}

	// Проверяем права пользователя
	if !claims.IsAdmin {
		log.Error("User is not administrator")
		return models.ErrPermissionDenied
	}

	// Пока что обновляем name
	// Можно полностью, когда будет доступен tokens-black-list
	if err := s.UserDal.UpdateUserName(user.Name, user.ID); err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return repo.ErrUserNotExist
		}
		log.Error("Failed to delete user data", "error", err)
		return models.ErrUnexpected
	}

	return nil
}
