package mock

import (
	"auth/internal/adapters/repo"
	"auth/internal/domain/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{}
}

func (*MockUserRepo) GetUser(email string) (models.User, error) {
	passHash, _ := bcrypt.GenerateFromPassword([]byte("validPassword"), bcrypt.DefaultCost)
	isAdmin := false
	switch email {
	case "adminEmail@gmail.com":
		isAdmin = true
	case "uniqueMail@gmail.com":
		return models.User{}, repo.ErrUserNotExist
	}

	user := models.User{
		ID:         1,
		Name:       "testName",
		Email:      email,
		IsAdmin:    isAdmin,
		Created_At: time.Now(),
		Updated_At: time.Now(),
	}
	user.SetPassword(string(passHash))
	return user, nil
}

func (*MockUserRepo) SaveUser(user *models.User) error {
	user.ID = 1
	return nil
}

func (*MockUserRepo) DeleteUser(userID int) error {
	return nil
}

func (*MockUserRepo) UpdateUserName(name string, userID int) error {
	return nil
}

func (*MockUserRepo) GetUserByID(userID int) (models.User, error) {
	return models.User{}, nil
}
