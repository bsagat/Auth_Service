package handler

import (
	"authService/internal/domain"
	"authService/internal/service"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
)

type AdminHandler struct {
	authServ  *service.AuthService
	adminServ *service.AdminService
	log       *slog.Logger
}

func NewAdminHandler(authServ *service.AuthService, adminServ *service.AdminService, log *slog.Logger) *AdminHandler {
	return &AdminHandler{
		authServ:  authServ,
		adminServ: adminServ,
		log:       log,
	}
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	adminToken, err := r.Cookie(domain.Access)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.log.Error("Failed to convert user id", "error", err)
		SendError(w, errors.New("user id is invalid"), http.StatusBadRequest)
		return
	}

	user, code, err := h.adminServ.GetUser(userID, adminToken.Value)
	if err != nil {
		h.log.Error("Failed to get user", "error", err)
		SendError(w, err, code)
		return
	}

	h.log.Info("User data fetch finished")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(&user); err != nil {
		h.log.Error("Failed to send user data", "error", err)
		SendError(w, errors.New("user data send error"), http.StatusInternalServerError)
		return
	}
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	adminToken, err := r.Cookie(domain.Access)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.log.Error("Failed to convert user id", "error", err)
		SendError(w, errors.New("user id is invalid"), http.StatusBadRequest)
		return
	}

	code, err := h.adminServ.DeleteUser(userID, adminToken.Value)
	if err != nil {
		h.log.Error("Failed to delete user", "error", err)
		SendError(w, err, code)
		return
	}

	h.log.Info("User deleted succesfully", "ID", userID)
	SendMessage(w, code, "User deleted succesfully")
}

func (h *AdminHandler) UpdateUserName(w http.ResponseWriter, r *http.Request) {
	adminToken, err := r.Cookie(domain.Access)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	var userReq UpdateUserReq
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		h.log.Error("Failed to decode json", "error", err)
		SendError(w, errors.New("invalid JSON data"), http.StatusBadRequest)
		return
	}

	code, err := h.adminServ.UpdateUser(domain.User{
		ID:   userReq.ID,
		Name: userReq.Name,
	}, adminToken.Value)
	if err != nil {
		h.log.Error("Failed to update user", "error", err)
		SendError(w, err, code)
		return
	}

	h.log.Info("User deleted succesfully", "ID", userReq.ID)
	SendMessage(w, code, "User updated succesfully")
}
