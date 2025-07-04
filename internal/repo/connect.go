package repo

import (
	"authService/internal/domain"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserDal struct {
	Db *sql.DB
}

func NewUserDal(Db *sql.DB) *UserDal {
	return &UserDal{Db: Db}
}

func Connect(config domain.DatabaseConf) (*sql.DB, error) {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", config.UserName, config.Password, config.Name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := migrateAdmin(db); err != nil {
		return nil, err
	}

	return db, nil
}

// Admin credentials registration
func migrateAdmin(Db *sql.DB) error {
	const op = "repo.migrateAdmin"

	admin := domain.User{
		Name:    os.Getenv("ADMIN_NAME"),
		Email:   os.Getenv("ADMIN_EMAIL"),
		IsAdmin: true,
	}
	admin.SetPassword(os.Getenv("ADMIN_PASSWORD"))

	var count int
	if err := Db.QueryRow(`SELECT COUNT(*) FROM Users;`).Scan(&count); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if count != 0 {
		slog.Warn("Table is not empty, skipping inizialization...")
		return nil
	}

	// Хешируем пароль
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(admin.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: failed to hash password: %w", op, err)
	}

	// Вставляем админа в таблицу
	_, err = Db.Exec(`
		INSERT INTO Users (Name, Email, Passhash, IsAdmin)
		VALUES ($1, $2, $3, $4);
	`, admin.Name, admin.Email, hashedPass, admin.IsAdmin)

	if err != nil {
		return fmt.Errorf("%s: failed to insert admin: %w", op, err)
	}

	slog.Info("Admin user created successfully", "name", admin.Name, "email", admin.Email)
	return nil
}
