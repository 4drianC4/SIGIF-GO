package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *CustomerHTTPHandler) ChangeStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
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
