package routers

import (
	"auth/internal/adapters/transport/http/dto"
	"auth/internal/domain/models"
	"auth/internal/service"
	"auth/pkg/utils"
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
	var user dto.LoginReq

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.log.Error("Failed to decode json", "error", err)
		utils.SendError(w, errors.New("invalid JSON data"), http.StatusBadRequest)
		return
	}

	// Валидация реквизитов
	if err := ValidateCredentials("valid Name", user.Email, user.Password, models.UserRole); err != nil {
		h.log.Error("Email address is invalid")
		utils.SendError(w, err, http.StatusBadRequest)
		return
	}

	tokens, err := h.authServ.Login(user.Email, user.Password)
	if err != nil {
		h.log.Error("Failed to auth user", "error", err)
		utils.SendError(w, err, utils.GetStatus(err))
		return
	}

	h.log.Info("User login finished")
	SetTokenCookies(w, tokens, r.TLS != nil)
	utils.SendMessage(w, http.StatusOK, "User login success")
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user dto.RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.log.Error("Failed to decode json", "error", err)
		utils.SendError(w, errors.New("invalid JSON data"), http.StatusBadRequest)
		return
	}

	// Валидация реквизитов
	if err := ValidateCredentials(user.Name, user.Email, user.Password, user.Role); err != nil {
		h.log.Error("Email address is invalid")
		utils.SendError(w, err, http.StatusBadRequest)
		return
	}

	userID, err := h.authServ.Register(user.Name, user.Email, user.Password, user.Role)
	if err != nil {
		h.log.Error("Failed to register user", "error", err)
		utils.SendError(w, err, utils.GetStatus(err))
		return
	}

	ClearTokenCookies(w)

	h.log.Info("User registered", "ID", userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(struct {
		UserID int `json:"ID"`
	}{
		UserID: userID,
	})
}

func (h *AuthHandler) CheckRole(w http.ResponseWriter, r *http.Request) {
	// Достаем access token
	tokenCookie, err := r.Cookie(models.Access)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		utils.SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	// Вызов основной логики
	existUser, err := h.authServ.RoleCheck(tokenCookie.Value)
	if err != nil {
		h.log.Error("Failed to check user role", "error", err)
		utils.SendError(w, err, utils.GetStatus(err))
		return
	}

	// Возвращаем ответ
	h.log.Info("User role check finished", "ID", existUser.ID, "is_admin", existUser.IsAdmin)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&existUser)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie(models.Refresh)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		utils.SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	tokens, err := h.tokenServ.Refresh(tokenCookie.Value)
	if err != nil {
		h.log.Error("Failed to refresh token", "error", err)
		utils.SendError(w, err, utils.GetStatus(err))
		return
	}

	h.log.Info("Token has been refreshed")
	SetTokenCookies(w, tokens, r.TLS != nil)
	w.WriteHeader(http.StatusOK)
}

func SetTokenCookies(w http.ResponseWriter, tokens models.TokenPair, hasTLS bool) {
	ClearTokenCookies(w)
	http.SetCookie(w, &http.Cookie{
		Name:     models.Access,
		Value:    tokens.AccessToken,
		Expires:  tokens.AccessExpiresAt.UTC(),
		HttpOnly: true,
		Secure:   hasTLS, // Отправка только через HTTPS (если передача зашифрованая >>> включить)
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     models.Refresh,
		Value:    tokens.RefreshToken,
		Expires:  tokens.RefreshExpiresAt.UTC(),
		HttpOnly: true,
		Secure:   hasTLS, // Отправка только через HTTPS (если передача зашифрованая >>> включить)
		SameSite: http.SameSiteStrictMode,
		Path:     "/refresh",
	})
}

func ClearTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   models.Access,
		Value:  "",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   models.Refresh,
		Value:  "",
		MaxAge: -1,
	})
}
