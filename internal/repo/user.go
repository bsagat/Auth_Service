package repo

import (
	"authService/internal/domain"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrUserNotExist = errors.New("user does not exist")
)

func (repo *UserDal) GetUser(email string) (domain.User, error) {
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
	var passHash string
	if err := repo.Db.QueryRow(query, email).
		Scan(&user.ID, &user.Name, &user.Email, &passHash, &user.IsAdmin, &user.Created_At, &user.Updated_At); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, fmt.Errorf("%s:%w", op, ErrUserNotExist)
		}
		return domain.User{}, fmt.Errorf("%s:%w", op, err)
	}
	user.SetPassword(passHash)

	return user, nil
}

// Saves user and sets his ID
func (repo *UserDal) SaveUser(user *domain.User) error {
	const op = "UserDal.SaveUser"
	query := `
	INSERT INTO Users (Name, Email, PassHash, IsAdmin)
	VALUES ($1, $2, $3, $4)
	RETURNING ID
	`

	// QueryRow для получения ID
	if err := repo.Db.QueryRow(query, user.Name, user.Email, user.GetPassword(), user.IsAdmin).
		Scan(&user.ID); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}
func (repo *UserDal) DeleteUser(userID int) error {
	const op = "UserDal.DeleteUser"
	query := `
		DELETE FROM Users
		WHERE ID = $1`

	res, err := repo.Db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s:%w", op, ErrUserNotExist)
	}

	return nil
}

func (repo *UserDal) UpdateUserName(name string, userID int) error {
	const op = "UserDal.UpdateUser"
	query := `UPDATE Users
	SET Name=$1 , Updated_at = Now()
	WHERE ID=$2
	`

	res, err := repo.Db.Exec(query, name, userID)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s:%w", op, ErrUserNotExist)
	}

	return nil
}
