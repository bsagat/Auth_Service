package dal

import (
	"authService/internal/domain"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrUserNotExist = errors.New("user does not exist")
)

func (dal *UserDal) GetUser(email string) (domain.User, error) {
	const op = "UserDal.GetUser"
	query := `
	SELECT 
		ID, Name, Email, PassHash, IsAdmin, Created_At, Coalesce(Updated_At,Created_At) 
	FROM   
		Users
	WHERE
		Email=$1
	LIMIT 
		1
	`

	var user domain.User
	if err := dal.Db.QueryRow(query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsAdmin, &user.Created_At, &user.Updated_At); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, fmt.Errorf("%s:%w", op, ErrUserNotExist)
		}
		return domain.User{}, fmt.Errorf("%s:%w", op, err)
	}

	return user, nil
}

// Saves user and sets his ID
func (dal *UserDal) SaveUser(user *domain.User) error {
	const op = "UserDal.SaveUser"
	query := `
	INSERT INTO Users (Name, Email, PassHash, IsAdmin)
	VALUES ($1, $2, $3, $4)
	RETURNING ID
	`

	if err := dal.Db.QueryRow(query, user.Name, user.Email, user.Password, user.IsAdmin).
		Scan(&user.ID); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}
