package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

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
