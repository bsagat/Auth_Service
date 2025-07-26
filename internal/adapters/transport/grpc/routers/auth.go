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

type AuthHandler struct {
	authServ  *service.AuthService
	tokenServ *service.TokenService
	log       *slog.Logger

	authv1.UnimplementedAuthServiceServer
}

func NewAuthHandler(authServ *service.AuthService, tokenServ *service.TokenService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authServ:  authServ,
		tokenServ: tokenServ,
		log:       log,
	}
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	email := req.GetEmail()
	password := req.GetPassword()

	// Валидация реквизитов
	if err := validate.Credentials("valid Name", email, password, models.UserRole); err != nil {
		h.log.Error("Credentials are invalid", "email", email, "password", password)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tokens, err := h.authServ.Login(email, password)
	if err != nil {
		h.log.Error("Failed to auth user", "error", err)
		return nil, status.Error(utils.GetGRPCStatus(err), err.Error())
	}

	h.log.Info("User login finished")
	return &authv1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	refresh := req.GetRefreshToken()

	tokens, err := h.tokenServ.Refresh(refresh)
	if err != nil {
		h.log.Error("Failed to refresh token", "error", err)
		return nil, status.Error(utils.GetGRPCStatus(err), err.Error())
	}

	h.log.Info("Token has been refreshed")
	return &authv1.RefreshResponse{
		NewAccessToken:  tokens.AccessToken,
		NewRefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	name, email, role, password := req.GetName(), req.GetEmail(), req.GetRole(), req.GetPassword()

	if err := validate.Credentials(name, email, password, role); err != nil {
		h.log.Error("Credentials are invalid")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := h.authServ.Register(name, email, password, role)
	if err != nil {
		h.log.Error("Failed to register user", "error", err)
		return nil, status.Error(utils.GetGRPCStatus(err), err.Error())
	}

	return &authv1.RegisterResponse{
		Id: int64(userID),
	}, nil
}

func (h *AuthHandler) WhoAmI(ctx context.Context, req *authv1.WhoAmIRequest) (*authv1.WhoAmIResponse, error) {
	token := req.GetToken()

	// Вызов основной логики
	existUser, err := h.authServ.RoleCheck(token)
	if err != nil {
		h.log.Error("Failed to check user role", "error", err)
		return nil, status.Error(utils.GetGRPCStatus(err), err.Error())
	}

	// Возвращаем ответ
	h.log.Info("User role check finished", "ID", existUser.ID, "is_admin", existUser.IsAdmin)
	return &authv1.WhoAmIResponse{
		User: &authv1.User{
			Id:        int64(existUser.ID),
			Name:      existUser.Name,
			Email:     existUser.Email,
			IsAdmin:   existUser.IsAdmin,
			CreatedAt: timestamppb.New(existUser.Created_At),
			UpdatedAt: timestamppb.New(existUser.Updated_At),
			Role:      existUser.Role,
		},
	}, nil
}
