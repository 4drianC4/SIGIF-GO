package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/response"
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
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	user, err := h.cmdHandler.HandleUpdate(c.UserContext(), command.UpdateUser{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Area:      req.Area,
	})
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
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if err := h.cmdHandler.HandleChangePassword(c.UserContext(), command.ChangePassword{
		ID:              id,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}
