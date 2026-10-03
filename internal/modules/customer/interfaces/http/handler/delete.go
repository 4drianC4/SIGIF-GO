package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// Delete maneja DELETE /customers/:id para aplicar baja lógica.
func (h *CustomerHTTPHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid customer id", 400))
	}

	tenantID := dtos.TenantIDFromContext(c)

	cmd := command.DeleteCustomer{
		ID:       id,
		TenantID: tenantID,
	}

	if err := h.cmdHandler.HandleDelete(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}
