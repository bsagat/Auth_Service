package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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
