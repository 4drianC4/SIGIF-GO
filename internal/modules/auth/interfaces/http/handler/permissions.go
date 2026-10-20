package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/application/dto"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *AuthHTTPHandler) MePermissions(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, sharedErrors.ErrUnauthorized)
	}

	user, permissions, err := h.cmdHandler.HandleMe(c.UserContext(), command.Me{UserID: userID})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if user == nil {
		return response.Error(c, fiber.StatusUnauthorized, sharedErrors.ErrUnauthorized)
	}

	return response.Success(c, dto.FromPermissions(user, permissions))
}
