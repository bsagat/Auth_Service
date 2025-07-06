package domain

type UserRepo interface {
	GetUser(email string) (User, error)
	SaveUser(user *User) error
	DeleteUser(userID int) error
	UpdateUserName(name string, userID int) error
}

type TokenService interface {
	GenerateTokens(user User) (TokenPair, error)
	Refresh(refreshToken string) (TokenPair, int, error)
	Validate(token string) (CustomClaims, error)
}
