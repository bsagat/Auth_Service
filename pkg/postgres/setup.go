package postgres

import (
	"auth/internal/domain/models"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type (
	DatabaseConf struct {
		Name     string `env:"DB_NAME"`
		Password string `env:"DB_PASSWORD"`
		Port     string `env:"DB_PORT"`
		UserName string `env:"DB_USER"`
	}

	AdminCredentials struct {
		Name     string `env:"ADMIN_NAME"`
		Password string `env:"ADMIN_PASSWORD"`
		Email    string `env:"ADMIN_EMAIL"`
	}
)

type PostgreDB struct {
	DB *sql.DB
}

// Осуществляет подключение к базе данных [postgres]
func Connect(cfg DatabaseConf, adminCreds AdminCredentials) (*PostgreDB, error) {
	const op = "postgres.Connect"
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", cfg.UserName, cfg.Password, cfg.Name)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := migrateAdmin(db, adminCreds); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &PostgreDB{DB: db}, nil
}

// проводит регистрацию main администратора
func migrateAdmin(Db *sql.DB, cred AdminCredentials) error {
	const op = "repo.migrateAdmin"

	admin := models.User{
		Name:    cred.Name,
		Email:   cred.Email,
		IsAdmin: true,
	}
	admin.SetPassword(cred.Password)

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
