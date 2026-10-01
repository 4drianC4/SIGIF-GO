package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/application/dto"
)

func (h *AuthCommandHandler) HandleRefresh(ctx context.Context, cmd command.Refresh) (*dto.TokenPair, error) {
	result, err := h.service.Refresh(ctx, cmd.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &dto.TokenPair{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
	}, nil
}
