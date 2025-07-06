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
	var user LoginReq
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

	h.log.Info("User login finished")
	SetTokenCookies(w, tokens)
	SendMessage(w, code, "User login success")
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user RegisterReq
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

	ClearTokenCookies(w)

	h.log.Info("User registered", "ID", userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(struct {
		UserID int `json:"ID"`
	}{
		UserID: userID,
	})
}

func (h *AuthHandler) CheckRole(w http.ResponseWriter, r *http.Request) {
	// Достаем access token
	tokenCookie, err := r.Cookie(domain.Access)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	// Вызов основной логики
	existUser, code, err := h.authServ.RoleCheck(tokenCookie.Value)
	if err != nil {
		h.log.Error("Failed to check user role", "error", err)
		SendError(w, err, code)
		return
	}

	// Возвращаем ответ
	h.log.Info("User role check finished", "ID", existUser.ID, "is_admin", existUser.IsAdmin)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&existUser)
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

	h.log.Info("Token has been refreshed")
	SetTokenCookies(w, tokens)
	w.WriteHeader(code)
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

func ClearTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   domain.Access,
		Value:  "",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   domain.Refresh,
		Value:  "",
		MaxAge: -1,
	})
}
