package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *AuthHTTPHandler) Logout(c *fiber.Ctx) error {
	var req dtos.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if err := h.cmdHandler.HandleLogout(c.UserContext(), command.Logout{RefreshToken: req.RefreshToken}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *AuthHTTPHandler) LogoutAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	uid, ok := userID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, sharedErrors.ErrUnauthorized)
	}

	if err := h.cmdHandler.HandleLogoutAll(c.UserContext(), command.LogoutAll{UserID: uid}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}
