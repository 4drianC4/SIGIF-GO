package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *UserHTTPHandler) Activate(c *fiber.Ctx) error {
	return h.changeStatus(c, entity.UserStatusActive)
}

func (h *UserHTTPHandler) Deactivate(c *fiber.Ctx) error {
	return h.changeStatus(c, entity.UserStatusInactive)
}

func (h *UserHTTPHandler) Suspend(c *fiber.Ctx) error {
	return h.changeStatus(c, entity.UserStatusSuspended)
}

func (h *UserHTTPHandler) changeStatus(c *fiber.Ctx, action entity.UserStatus) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if err := h.cmdHandler.HandleChangeStatus(c.UserContext(), command.ChangeStatus{ID: id, Action: action}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}