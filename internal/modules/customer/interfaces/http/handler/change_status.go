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
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	companyID, err := companyIDFromContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.ChangeCustomerStatus{
		ID:        id,
		CompanyID: companyID,
		Status:    dtos.StatusFromRequest(req.Status),
	}

	customer, err := h.cmdHandler.HandleChangeStatus(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromEntity(customer))
}
