package utils

import (
	"auth/internal/adapters/repo"
	"auth/internal/domain/models"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"google.golang.org/grpc/codes"
)

func ShowHelp() {
	text :=
		`Auth Service
Flags:		
	--help 	   [ Shows help message ]
	--port     [ Default auth service port number ]
	--host     [ Default auth service host settings ]
	--env      [ Application environment: local | dev | prod ]`
	fmt.Println(text)
	os.Exit(0)
}

func SendMessage(w http.ResponseWriter, code int, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(&struct {
		Message string `json:"message"`
	}{
		Message: message,
	}); err != nil {
		return err
	}
	return nil
}

func SendError(w http.ResponseWriter, err error, code int) error {
	errMessage := struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(&errMessage); err != nil {
		slog.Error("Failed to send error", "error", err)
		return err
	}
	return nil
}

func GetHTTpStatus(err error) int {
	switch {
	case errors.Is(err, models.ErrInvalidToken), errors.Is(err, models.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, models.ErrPermissionDenied):
		return http.StatusForbidden
	case errors.Is(err, repo.ErrUserNotExist):
		return http.StatusNotFound
	case errors.Is(err, models.ErrNotUniqueEmail), errors.Is(err, models.ErrCannotDeleteSelf):
		return http.StatusConflict
	case errors.Is(err, models.ErrCannotCreateAdmin), errors.Is(err, models.ErrCannotDeleteSelf):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func GetGRPCStatus(err error) codes.Code {
	switch {
	case errors.Is(err, models.ErrInvalidToken), errors.Is(err, models.ErrInvalidCredentials):
		return codes.Unauthenticated
	case errors.Is(err, models.ErrPermissionDenied):
		return codes.PermissionDenied
	case errors.Is(err, repo.ErrUserNotExist):
		return codes.NotFound
	case errors.Is(err, models.ErrNotUniqueEmail):
		return codes.AlreadyExists
	case errors.Is(err, models.ErrCannotCreateAdmin), errors.Is(err, models.ErrCannotDeleteSelf):
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}
