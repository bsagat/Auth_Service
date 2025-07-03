package handler

import (
	"authService/internal/domain"
	"authService/internal/service"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type AuthHandler struct {
	authServ  *service.AuthService
	tokenServ *service.TokenService
	log       *slog.Logger
}

func NewAuthHandler(authServ *service.AuthService, tokenServ *service.TokenService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authServ:  authServ,
		tokenServ: tokenServ,
		log:       log,
	}
}

// Возвращает токен
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.log.Error("Failed to decode json", "error", err)
		SendError(w, errors.New("invalid JSON data"), http.StatusBadRequest)
		return
	}

	tokens, code, err := h.authServ.Login(user.Email, user.Password)
	if err != nil {
		h.log.Error("Failed to auth user", "error", err)
		SendError(w, err, code)
		return
	}

	SetTokenCookies(w, tokens)
	w.WriteHeader(code)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.log.Error("Failed to decode json", "error", err)
		SendError(w, errors.New("invalid JSON data"), http.StatusBadRequest)
		return
	}

	userID, code, err := h.authServ.Register(user.Name, user.Email, user.Password)
	if err != nil {
		h.log.Error("Failed to register user", "error", err)
		SendError(w, err, code)
		return
	}

	h.log.Info("User registered", "user_id", userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(struct {
		UserID int `json:"user_id"`
	}{
		UserID: userID,
	})
}

func (h *AuthHandler) IsAdmin(w http.ResponseWriter, r *http.Request) {
	// Достаем access token
	tokenCookie, err := r.Cookie(domain.Access)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	isAdmin, code, err := h.authServ.IsAdmin(tokenCookie.Value)
	if err != nil {
		h.log.Error("Failed to check user role", "error", err)
		SendError(w, err, code)
		return
	}

	// Возвращаем ответ
	h.log.Info("User role check finished", "is_admin", isAdmin)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(struct {
		IsAdmin bool `json:"is_admin"`
	}{
		IsAdmin: isAdmin,
	})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie(domain.Refresh)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	tokens, code, err := h.tokenServ.Refresh(tokenCookie.Value)
	if err != nil {
		h.log.Error("Failed to refresh token", "error", err)
		SendError(w, err, code)
		return
	}

	SetTokenCookies(w, tokens)
	w.WriteHeader(code)
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

func SetTokenCookies(w http.ResponseWriter, tokens domain.TokenPair) {
	http.SetCookie(w, &http.Cookie{
		Name:     domain.Access,
		Value:    tokens.AccessToken,
		Expires:  tokens.AccessExpiresAt,
		HttpOnly: true,
		Secure:   false, // Отправка только через HTTPS (если передача зашифрованая >>> включить)
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     domain.Refresh,
		Value:    tokens.RefreshToken,
		Expires:  tokens.RefreshExpiresAt,
		HttpOnly: true,
		Secure:   false, // Отправка только через HTTPS (если передача зашифрованая >>> включить)
		SameSite: http.SameSiteStrictMode,
		Path:     "/refresh",
	})
}
