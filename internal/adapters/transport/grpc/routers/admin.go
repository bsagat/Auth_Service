package routers

import (
	validate "auth/internal/adapters/transport"
	authv1 "auth/internal/adapters/transport/grpc/gen"
	"auth/internal/domain/models"
	"auth/internal/service"
	"auth/pkg/utils"
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AdminHandler struct {
	authServ  *service.AuthService
	adminServ *service.AdminService
	log       *slog.Logger

	authv1.UnimplementedAdminServiceServer
}

func NewAdminHandler(authServ *service.AuthService, adminServ *service.AdminService, log *slog.Logger) *AdminHandler {
	return &AdminHandler{
		authServ:  authServ,
		adminServ: adminServ,
		log:       log,
	}
}

func (h *AdminHandler) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.GetUserResponse, error) {
	adminToken := req.GetAdminToken()
	userID := req.GetUserId()

	if userID == 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID is empty")
	}

	user, err := h.adminServ.GetUser(int(userID), adminToken)
	if err != nil {
		h.log.Error("Failed to get user", "error", err)
		return nil, status.Errorf(utils.GetGRPCStatus(err), "failed to get user data: %v", err)
	}

	h.log.Info("User data fetch finished")
	return &authv1.GetUserResponse{
		User: &authv1.User{
			Id:        userID,
			Name:      user.Name,
			Email:     user.Email,
			IsAdmin:   user.IsAdmin,
			CreatedAt: timestamppb.New(user.Created_At),
			UpdatedAt: timestamppb.New(user.Updated_At),
			Role:      user.Role,
		},
	}, nil
}

func (h *AdminHandler) UpdateUser(ctx context.Context, req *authv1.UpdateRequest) (*authv1.UpdateResponse, error) {
	adminToken := req.GetAdminToken()
	userID := int(req.GetUserId())
	role := req.GetRole()
	name := req.GetName()

	// Валидируем запрос
	if err := validate.UserReq(userID, name, role); err != nil {
		h.log.Error("Update request is invalid", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "update request is invalid: %v", err)
	}

	if err := h.adminServ.UpdateUser(models.User{
		ID:   userID,
		Name: name,
		Role: role,
	}, adminToken); err != nil {
		h.log.Error("Failed to update user", "error", err)
		return nil, status.Errorf(utils.GetGRPCStatus(err), "failed to update user data: %v", err)
	}

	h.log.Info("User updated succesfully", "ID", userID)
	return &authv1.UpdateResponse{
		Message: "User updated succesfully",
	}, nil
}

func (h *AdminHandler) DeleteUser(ctx context.Context, req *authv1.DeleteRequest) (*authv1.DeleteResponse, error) {
	adminToken := req.GetAdminToken()
	userID := req.GetUserId()

	if userID == 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID is empty")
	}

	if err := h.adminServ.DeleteUser(int(userID), adminToken); err != nil {
		h.log.Error("Failed to delete user", "error", err)
		return nil, status.Errorf(utils.GetGRPCStatus(err), "failed to delete user data: %v", err)
	}

	h.log.Info("User deleted succesfully", "ID", userID)
	return &authv1.DeleteResponse{
		Message: "User deleted succesfully",
	}, nil
}
