package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/response"
	sharedValidator "github.com/sigif/sigif-go/internal/shared/validator"
)

func (h *UserHTTPHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	var req dtos.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := sharedValidator.New().Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.UpdateUser{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		AvatarURL: req.AvatarURL,
		Roles:     req.Roles,
		Settings:  req.Settings,
	}

	user, err := h.cmdHandler.HandleUpdate(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dtos.ToResponse(user))
}

func (h *UserHTTPHandler) ChangePassword(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	var req dtos.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := sharedValidator.New().Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.ChangePassword{
		ID:              id,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := h.cmdHandler.HandleChangePassword(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}