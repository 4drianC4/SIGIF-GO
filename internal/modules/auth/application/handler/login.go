package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/application/dto"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/service"
	"github.com/sigif/sigif-go/internal/shared/jwt"
)

func (h *AuthCommandHandler) HandleLogin(ctx context.Context, cmd command.Login) (dto.AuthResponse, error) {
	result, err := h.service.Login(ctx, service.LoginParams{
		TenantID:  cmd.TenantID,
		Email:     cmd.Email,
		Password:  cmd.Password,
		UserAgent: cmd.UserAgent,
		IPAddress: cmd.IPAddress,
	})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	tokenPair := &jwt.TokenPair{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
	}
	return dto.FromUser(result.User, tokenPair), nil
}
