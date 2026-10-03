package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// ChangeStatus maneja PATCH /customers/:id/status.
func (h *CustomerHTTPHandler) ChangeStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid customer id", 400))
	}

	var req dtos.ChangeStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if errs := h.validator.Validate(req); errs != nil {
		return response.Error(c, fiber.StatusBadRequest, errs)
	}

	tenantID := dtos.TenantIDFromContext(c)

	cmd := command.ChangeCustomerStatus{
		ID:       id,
		TenantID: tenantID,
		Status:   dtos.StatusFromRequest(req.Status),
	}

	customer, err := h.cmdHandler.HandleChangeStatus(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	customerDTO := dto.FromEntity(customer)
	return response.Success(c, dtos.ToResponse(customerDTO))
}
