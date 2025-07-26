package routers

import (
	"auth/internal/adapters/transport/http/dto"
	"auth/internal/domain/models"
	"auth/internal/service"
	"auth/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
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
	adminToken, err := r.Cookie(models.Access)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		utils.SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.log.Error("Failed to convert user id", "error", err)
		utils.SendError(w, errors.New("user id is invalid"), http.StatusBadRequest)
		return
	}

	user, err := h.adminServ.GetUser(userID, adminToken.Value)
	if err != nil {
		h.log.Error("Failed to get user", "error", err)
		utils.SendError(w, err, utils.GetStatus(err))
		return
	}

	h.log.Info("User data fetch finished")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(struct {
		User     models.User `json:"user"`
		Password string      `json:"password"`
	}{
		User:     user,
		Password: user.GetPassword(),
	}); err != nil {
		h.log.Error("Failed to send user data", "error", err)
		utils.SendError(w, errors.New("user data send error"), http.StatusInternalServerError)
		return
	}
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	adminToken, err := r.Cookie(models.Access)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		utils.SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.log.Error("Failed to convert user id", "error", err)
		utils.SendError(w, errors.New("user id is invalid"), http.StatusBadRequest)
		return
	}

	if err := h.adminServ.DeleteUser(userID, adminToken.Value); err != nil {
		h.log.Error("Failed to delete user", "error", err)
		utils.SendError(w, err, utils.GetStatus(err))
		return
	}

	h.log.Info("User deleted succesfully", "ID", userID)
	utils.SendMessage(w, http.StatusNoContent, "User deleted succesfully")
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	adminToken, err := r.Cookie(models.Access)
	if err != nil {
		h.log.Error("Failed to get cookie", "error", err)
		utils.SendError(w, errors.New("cookie not found"), http.StatusUnauthorized)
		return
	}

	var userReq dto.UpdateUserReq
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		h.log.Error("Failed to decode json", "error", err)
		utils.SendError(w, errors.New("invalid JSON data"), http.StatusBadRequest)
		return
	}

	// Валидируем запрос
	if err := ValidateUserReq(userReq); err != nil {
		h.log.Error("Update request is invalid", "error", err)
		utils.SendError(w, fmt.Errorf("update request is invalid: %w", err), http.StatusBadRequest)
		return
	}

	if err := h.adminServ.UpdateUser(models.User{
		ID:   userReq.ID,
		Name: userReq.Name,
		Role: userReq.Role,
	}, adminToken.Value); err != nil {
		h.log.Error("Failed to update user", "error", err)
		utils.SendError(w, err, utils.GetStatus(err))
		return
	}

	h.log.Info("User deleted succesfully", "ID", userReq.ID)
	utils.SendMessage(w, http.StatusOK, "User updated succesfully")
}
