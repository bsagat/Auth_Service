package repo

import (
	"auth/internal/domain/models"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrUserNotExist = errors.New("user does not exist")
)

type UserDal struct {
	Db *sql.DB
}

func NewUserDal(Db *sql.DB) *UserDal {
	return &UserDal{Db: Db}
}

func (repo *UserDal) GetUser(email string) (models.User, error) {
	const op = "UserDal.GetUser"
	query := `
	SELECT 
		ID, Name, Email, PassHash, IsAdmin, Created_At, Coalesce(Updated_At,Created_At), Role 
	FROM   
		Users
	WHERE
		Email=$1
	LIMIT 
		1
	`

	var user models.User
	var passHash string
	if err := repo.Db.QueryRow(query, email).
		Scan(&user.ID, &user.Name, &user.Email, &passHash, &user.IsAdmin, &user.Created_At, &user.Updated_At, &user.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s:%w", op, ErrUserNotExist)
		}
		return models.User{}, fmt.Errorf("%s:%w", op, err)
	}
	user.SetPassword(passHash)

	return user, nil
}

func (repo *UserDal) GetUserByID(userID int) (models.User, error) {
	const op = "UserDal.GetUser"
	query := `
	SELECT 
		ID, Name, Email, PassHash, IsAdmin, Created_At, Coalesce(Updated_At,Created_At), Role 
	FROM   
		Users
	WHERE
		ID=$1
	LIMIT 
		1
	`

	var user models.User
	var passHash string
	if err := repo.Db.QueryRow(query, userID).
		Scan(&user.ID, &user.Name, &user.Email, &passHash, &user.IsAdmin, &user.Created_At, &user.Updated_At, &user.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s:%w", op, ErrUserNotExist)
		}
		return models.User{}, fmt.Errorf("%s:%w", op, err)
	}
	user.SetPassword(passHash)

	return user, nil
}

// Saves user and sets his ID
func (repo *UserDal) SaveUser(user *models.User) error {
	const op = "UserDal.SaveUser"
	query := `
	INSERT INTO Users (Name, Email, PassHash, IsAdmin, Role)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING ID
	`

	// QueryRow для получения ID
	if err := repo.Db.QueryRow(query, user.Name, user.Email, user.GetPassword(), user.IsAdmin, user.Role).
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

func (repo *UserDal) UpdateUser(name string, role string, userID int) error {
	const op = "UserDal.UpdateUser"
	query := `UPDATE Users
	SET Name=$1 , Role = $2 , Updated_at = Now()
	WHERE ID=$3
	`

	res, err := repo.Db.Exec(query, name, role, userID)
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
