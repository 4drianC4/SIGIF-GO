package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *CustomerHTTPHandler) Create(c *fiber.Ctx) error {
	var req dtos.CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if errs := h.validator.Validate(req); errs != nil {
		return response.Error(c, fiber.StatusBadRequest, errs)
	}

	tenantID := dtos.TenantIDFromContext(c)
	if tenantID.String() == "00000000-0000-0000-0000-000000000000" {
		return response.Error(
			c,
			fiber.StatusBadRequest,
			sharedErrors.New(sharedErrors.CodeBadRequest, "tenant is required", fiber.StatusBadRequest),
		)
	}

	cmd := command.CreateCustomer{
		TenantID:       tenantID,
		LegalName:      req.LegalName,
		DocumentType:   dtos.DocumentTypeFromRequest(req.DocumentType),
		DocumentNumber: req.DocumentNumber,
		Phone:          req.Phone,
		Email:          req.Email,
	}

	customer, err := h.cmdHandler.HandleCreate(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	customerDTO := dto.FromEntity(customer)
	return response.Created(c, dtos.ToResponse(customerDTO))
}
