package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *CustomerHTTPHandler) Create(c *fiber.Ctx) error {
	var req dtos.CreateCustomerRequest
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

	cmd := command.CreateCustomer{
		CompanyID:      companyID,
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

	return response.Created(c, dto.FromEntity(customer))
}
