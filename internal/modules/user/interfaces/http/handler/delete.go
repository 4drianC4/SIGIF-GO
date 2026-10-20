package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *UserHTTPHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if err := h.cmdHandler.HandleDelete(c.UserContext(), command.DeleteUser{ID: id}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *UserHTTPHandler) Activate(c *fiber.Ctx) error {
	return h.changeStatus(c, true)
}

func (h *UserHTTPHandler) Deactivate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if err := h.cmdHandler.HandleChangeStatus(c.UserContext(), command.ChangeStatus{ID: id, Active: false}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	user, err := h.queryHandler.HandleGet(c.UserContext(), query.GetUser{ID: id})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dtos.ToResponse(user))
}

func (h *UserHTTPHandler) changeStatus(c *fiber.Ctx, active bool) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if err := h.cmdHandler.HandleChangeStatus(c.UserContext(), command.ChangeStatus{ID: id, Active: active}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}
