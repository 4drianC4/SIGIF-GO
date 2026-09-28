package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/application/handler"
	"github.com/sigif/sigif-go/internal/modules/auth/application/dto"
	"github.com/sigif/sigif-go/internal/shared/response"
)

type AuthHTTPHandler struct {
	cmdHandler *handler.AuthCommandHandler
}

func NewAuthHTTPHandler(cmdHandler *handler.AuthCommandHandler) *AuthHTTPHandler {
	return &AuthHTTPHandler{cmdHandler: cmdHandler}
}

func (h *AuthHTTPHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.LoginCommand{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: c.Get("User-Agent"),
		IPAddress: c.IP(),
	}

	result, err := h.cmdHandler.HandleLogin(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, result)
}

func (h *AuthHTTPHandler) Refresh(c *fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.RefreshCommand{RefreshToken: req.RefreshToken}
	result, err := h.cmdHandler.HandleRefresh(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, result)
}

func (h *AuthHTTPHandler) Logout(c *fiber.Ctx) error {
	var req dto.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.LogoutCommand{RefreshToken: req.RefreshToken}
	if err := h.cmdHandler.HandleLogout(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *AuthHTTPHandler) LogoutAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	cmd := command.LogoutAllCommand{UserID: userID}
	if err := h.cmdHandler.HandleLogoutAll(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}