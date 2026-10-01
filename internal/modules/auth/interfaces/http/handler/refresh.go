package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *AuthHTTPHandler) Refresh(c *fiber.Ctx) error {
	var req dtos.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	result, err := h.cmdHandler.HandleRefresh(c.UserContext(), command.Refresh{RefreshToken: req.RefreshToken})
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, err)
	}

	return response.Success(c, result)
}
