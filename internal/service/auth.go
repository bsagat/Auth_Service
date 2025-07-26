package service

import (
	"auth/internal/adapters/repo"
	"auth/internal/domain/models"
	"auth/internal/domain/ports"
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserDal   ports.UserRepo
	TokenServ ports.TokenService
	log       *slog.Logger
}

func NewAuthService(UserDal ports.UserRepo, TokenServ ports.TokenService, log *slog.Logger) *AuthService {
	return &AuthService{
		UserDal:   UserDal,
		TokenServ: TokenServ,
		log:       log,
	}
}

// Returns (AccessToken, RefreshToken, statusCode, error message)
func (s *AuthService) Login(email, password string) (models.TokenPair, error) {
	const op = "AuthService.Login"
	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("User login started")

	// Проверяем существует ли пользователь
	existUser, err := s.UserDal.GetUser(email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return models.TokenPair{}, repo.ErrUserNotExist
		}
		log.Error("Failed to check user uniqueness", "error", err)
		return models.TokenPair{}, models.ErrUnexpected
	}

	// Сверяем пароли user-a и existing user-s с помощью compareHash
	if err := bcrypt.CompareHashAndPassword([]byte(existUser.GetPassword()), []byte(password)); err != nil {
		log.Error("Invalid credentials", "error", err)
		return models.TokenPair{}, models.ErrInvalidCredentials
	}

	// Генерируем токены
	tokens, err := s.TokenServ.GenerateTokens(existUser)
	if err != nil {
		log.Error("Failed to generate token", "error", err)
		return models.TokenPair{}, models.ErrTokenGenerateFail
	}

	return tokens, nil
}

func (s *AuthService) Register(name, email, password, role string) (int, error) {
	const op = "AuthService.Register"
	log := s.log.With(
		slog.String("op", op),
		slog.String("name", name),
		slog.String("email", email),
	)
	log.Info("User register started")

	// Проверяем уникальный ли email
	if user, err := s.UserDal.GetUser(email); err != nil && !errors.Is(err, repo.ErrUserNotExist) {
		log.Error("Failed to check user uniqueness", "error", err)
		return 0, models.ErrUnexpected
	} else {
		if user.ID != 0 {
			log.Error("User email is not unique")
			return 0, models.ErrNotUniqueEmail
		}
	}

	// Генерация хэша с defaultSolt(чем оно выше, тем лучше защищен хэш)
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to generate hash from password", "error", err)
		return 0, models.ErrUnexpected
	}

	if role == models.AdminRole {
		log.Error("Attempt to create admin via API")
		return 0, models.ErrCannotCreateAdmin
	}

	// Сохраняем нового пользователя
	newUser := models.User{
		Name:  name,
		Email: email,
		Role:  role,
	}
	newUser.SetPassword(string(hashedPass))

	if err = s.UserDal.SaveUser(&newUser); err != nil {
		log.Error("Failed to save user", "error", err)
		return 0, models.ErrUnexpected
	}

	return newUser.ID, nil
}

func (s *AuthService) RoleCheck(token string) (models.User, error) {
	const op = "AuthService.IsAdmin"
	log := s.log.With(
		slog.String("op", op),
	)
	log.Info("Role check started")

	// Валидируем его
	claim, err := s.TokenServ.Validate(token)
	if err != nil {
		log.Error("Access token is invalid", "error", err)
		return models.User{}, models.ErrInvalidToken
	}

	// Проверяем существует ли пользователь
	existUser, err := s.UserDal.GetUser(claim.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return models.User{}, repo.ErrUserNotExist
		}
		log.Error("Failed to check user uniqueness", "error", err)
		return models.User{}, models.ErrUnexpected
	}

	// Читаем админ ли он
	return existUser, nil
}
