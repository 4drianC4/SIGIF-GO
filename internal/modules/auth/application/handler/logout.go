package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
)

func (h *AuthCommandHandler) HandleLogout(ctx context.Context, cmd command.Logout) error {
	return h.service.Logout(ctx, cmd.RefreshToken)
}

func (h *AuthCommandHandler) HandleLogoutAll(ctx context.Context, cmd command.LogoutAll) error {
	return h.service.LogoutAll(ctx, cmd.UserID)
}
