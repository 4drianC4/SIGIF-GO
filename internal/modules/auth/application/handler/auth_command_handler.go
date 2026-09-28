package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/application/dto"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/service"
)

type AuthCommandHandler struct {
	service *service.AuthService
}

func NewAuthCommandHandler(service *service.AuthService) *AuthCommandHandler {
	return &AuthCommandHandler{service: service}
}

func (h *AuthCommandHandler) HandleLogin(ctx context.Context, cmd command.LoginCommand) (dto.AuthResponse, error) {
	tokens, user, err := h.service.Login(ctx, cmd.Email, cmd.Password, cmd.UserAgent, cmd.IPAddress)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return dto.ToAuthResponse(user, tokens), nil
}

func (h *AuthCommandHandler) HandleRefresh(ctx context.Context, cmd command.RefreshCommand) (dto.TokenPairResponse, error) {
	tokens, err := h.service.Refresh(ctx, cmd.RefreshToken)
	if err != nil {
		return dto.TokenPairResponse{}, err
	}

	return dto.ToTokenPairResponse(tokens), nil
}

func (h *AuthCommandHandler) HandleLogout(ctx context.Context, cmd command.LogoutCommand) error {
	return h.service.Logout(ctx, cmd.RefreshToken)
}

func (h *AuthCommandHandler) HandleLogoutAll(ctx context.Context, cmd command.LogoutAllCommand) error {
	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return err
	}
	return h.service.LogoutAll(ctx, userID)
}