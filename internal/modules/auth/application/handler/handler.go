package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/application/dto"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/service"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type AuthCommandHandler struct {
	service *service.AuthService
}

func NewAuthCommandHandler(service *service.AuthService) *AuthCommandHandler {
	return &AuthCommandHandler{service: service}
}

func (h *AuthCommandHandler) HandleLogin(ctx context.Context, cmd command.Login) (dto.AuthResponse, error) {
	result, err := h.service.Login(ctx, service.LoginParams{
		Email:     cmd.Email,
		Password:  cmd.Password,
		IPAddress: cmd.IPAddress,
		Device:    cmd.Device,
	})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return dto.AuthResponse{
		User: dto.FromUser(result.User, result.Permissions),
		Token: dto.Token{
			AccessToken: result.AccessToken,
			ExpiresIn:   result.ExpiresIn,
			TokenType:   result.TokenType,
		},
	}, nil
}

func (h *AuthCommandHandler) HandleLogout(ctx context.Context, cmd command.Logout) error {
	return h.service.Logout(ctx, cmd.TokenHash)
}

func (h *AuthCommandHandler) HandleMe(ctx context.Context, cmd command.Me) (*userEntity.AppUser, []string, error) {
	return h.service.Me(ctx, cmd.UserID)
}
