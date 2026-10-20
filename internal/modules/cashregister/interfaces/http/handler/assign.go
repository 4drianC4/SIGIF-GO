package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/command"
	"github.com/sigif/sigif-go/internal/modules/cashregister/application/dto"
	"github.com/sigif/sigif-go/internal/modules/cashregister/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// Assign gives the selected cash register to the authenticated cashier
// (HU-10-01).
func (h *CashRegisterHTTPHandler) Assign(c *fiber.Ctx) error {
	var req dtos.AssignCashRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid request body", 400))
	}
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	userID, companyID, err := actorFromContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	assignment, err := h.cmdHandler.HandleAssign(c.UserContext(), command.AssignCashRegister{
		CompanyID:      companyID,
		UserID:         userID,
		CashRegisterID: uuid.MustParse(req.CashRegisterID),
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Created(c, dto.FromAssignmentView(assignment))
}
