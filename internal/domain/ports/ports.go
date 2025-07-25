package ports

import "auth/internal/domain/models"

type UserRepo interface {
	GetUser(email string) (models.User, error)
	GetUserByID(userID int) (models.User, error)
	SaveUser(user *models.User) error
	DeleteUser(userID int) error
	UpdateUserName(name string, userID int) error
}

type TokenService interface {
	GenerateTokens(user models.User) (models.TokenPair, error)
	Refresh(refreshToken string) (models.TokenPair, error)
	Validate(token string) (models.CustomClaims, error)
}
