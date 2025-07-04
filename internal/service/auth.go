package service

import (
	"authService/internal/domain"
	"authService/internal/repo"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmptyEmail         = errors.New("email field is required")
	ErrEmptyPassword      = errors.New("password field is required")
	ErrEmptyName          = errors.New("name field is required")
	ErrInvalidCredentials = errors.New("credentials are invalid")
	ErrInvalidEmail       = errors.New("email is invalid")
	ErrInvalidPassword    = errors.New("password must be in range of 8 and 72 bytes")
	ErrInvalidName        = errors.New("name must be in range of 4 and 72 bytes")
)

type AuthService struct {
	UserDal   *repo.UserDal
	TokenServ *TokenService
	log       *slog.Logger
}

func NewAuthService(UserDal *repo.UserDal, TokenServ *TokenService, log *slog.Logger) *AuthService {
	return &AuthService{
		UserDal:   UserDal,
		TokenServ: TokenServ,
		log:       log,
	}
}

// Returns (AccessToken, RefreshToken, statusCode, error message)
func (s *AuthService) Login(email, password string) (domain.TokenPair, int, error) {
	const op = "AuthService.Login"
	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("User login started")

	// Валидация логин реквизитов
	if err := ValidateLogin(email, password); err != nil {
		log.Error("Invalid login credentials", "error", err)
		return domain.TokenPair{}, http.StatusBadRequest, err
	}

	// Проверяем существует ли пользователь
	existUser, err := s.UserDal.GetUser(email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return domain.TokenPair{}, http.StatusUnauthorized, repo.ErrUserNotExist
		}
		log.Error("Failed to check user uniqueness", "error", err)
		return domain.TokenPair{}, http.StatusInternalServerError, errors.New("failed to check user uniqueness")
	}

	// Сверяем пароли user-a и existing user-s с помощью compareHash
	if err := bcrypt.CompareHashAndPassword([]byte(existUser.GetPassword()), []byte(password)); err != nil {
		log.Error("Invalid credentials", "error", err)
		return domain.TokenPair{}, http.StatusUnauthorized, ErrInvalidCredentials
	}

	// Генерируем токены
	tokens, err := s.TokenServ.GenerateTokens(existUser)
	if err != nil {
		log.Error("Failed to generate token", "error", err)
		return domain.TokenPair{}, http.StatusInternalServerError, err
	}

	return tokens, http.StatusOK, nil
}

func (s *AuthService) Register(name, email, password string) (int, int, error) {
	const op = "AuthService.Register"
	log := s.log.With(
		slog.String("op", op),
		slog.String("name", name),
		slog.String("email", email),
	)
	log.Info("User register started")

	// Валидация реквизитов
	if err := ValidateRegister(name, email, password); err != nil {
		log.Error("Email address is invalid")
		return 0, http.StatusBadRequest, err
	}

	// Проверяем уникальный ли email
	if user, err := s.UserDal.GetUser(email); err != nil && !errors.Is(err, repo.ErrUserNotExist) {
		log.Error("Failed to check user uniqueness", "error", err)
		return 0, http.StatusInternalServerError, errors.New("failed to check user uniqueness")
	} else {
		if user.ID != 0 {
			log.Error("User email is not unique")
			return 0, http.StatusConflict, errors.New("email is not unique")
		}
	}

	// Генерация хэша с defaultSolt(чем оно выше, тем лучше защищен хэш)
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to generate hash from password", "error", err)
		return 0, http.StatusInternalServerError, errors.New("failed to generate hash")
	}

	// Сохраняем нового пользователя
	newUser := domain.User{
		Name:  name,
		Email: email,
	}
	newUser.SetPassword(string(hashedPass))

	if err = s.UserDal.SaveUser(&newUser); err != nil {
		log.Error("Failed to save user", "error", err)
		return 0, http.StatusInternalServerError, errors.New("failed to save user")
	}

	return newUser.ID, http.StatusOK, nil
}

func (s *AuthService) IsAdmin(token string) (bool, int, error) {
	const op = "AuthService.IsAdmin"
	log := s.log.With(
		slog.String("op", op),
	)
	log.Info("Role check started")

	// Валидируем его
	claim, err := s.TokenServ.Validate(token)
	if err != nil {
		log.Error("Access token is invalid", "error", err)
		return false, http.StatusUnauthorized, err
	}

	// Проверяем существует ли пользователь
	existUser, err := s.UserDal.GetUser(claim.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			log.Error("User is not exist")
			return false, http.StatusUnauthorized, repo.ErrUserNotExist
		}
		log.Error("Failed to check user uniqueness", "error", err)
		return false, http.StatusInternalServerError, errors.New("failed to check user uniqueness")
	}

	// Читаем админ ли он
	return existUser.IsAdmin, http.StatusOK, nil
}

func ValidateLogin(email, password string) error {
	if len(email) == 0 {
		return ErrEmptyEmail
	}

	if len(password) == 0 {
		return ErrEmptyPassword
	}

	return nil
}

func ValidateRegister(name, email, password string) error {
	// Name check
	if len(name) == 0 {
		return ErrEmptyName
	}

	if len(name) < 4 || len(name) > 72 {
		return ErrInvalidName
	}

	// Email check
	if len(email) == 0 || len(email) > 255 {
		return ErrEmptyEmail
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	// Password check
	if len(password) == 0 {
		return ErrEmptyPassword
	}

	if len(password) > 72 || len(password) < 8 {
		return ErrInvalidPassword
	}

	return nil
}
