package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *AuthHTTPHandler) Login(c *fiber.Ctx) error {
	var req dtos.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	result, err := h.cmdHandler.HandleLogin(c.UserContext(), command.Login{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.IP(),
		Device:    c.Get("User-Agent"),
	})
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, err)
	}

	return response.Success(c, result)
}
