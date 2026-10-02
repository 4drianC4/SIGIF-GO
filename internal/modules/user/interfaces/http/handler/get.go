package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *UserHTTPHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	user, err := h.queryHandler.HandleGet(c.UserContext(), query.GetUser{ID: id})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if user == nil {
		return response.Error(c, fiber.StatusNotFound, sharedErrors.ErrNotFound)
	}

	return response.Success(c, dtos.ToResponse(user))
}

func (h *UserHTTPHandler) GetByEmail(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}

	tenantID := dtos.TenantIDFromContext(c)
	user, err := h.queryHandler.HandleGetByEmail(c.UserContext(), query.GetUserByEmail{
		TenantID: tenantID,
		Email:    email,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if user == nil {
		return response.Error(c, fiber.StatusNotFound, sharedErrors.ErrNotFound)
	}

	return response.Success(c, dtos.ToResponse(user))
}