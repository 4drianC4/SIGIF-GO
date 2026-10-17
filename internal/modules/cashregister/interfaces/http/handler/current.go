package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/dto"
	"github.com/sigif/sigif-go/internal/modules/cashregister/application/query"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// Current returns the cash register held by the authenticated cashier
// (HU-10-01).
func (h *CashRegisterHTTPHandler) Current(c *fiber.Ctx) error {
	userID, companyID, err := actorFromContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	assignment, err := h.queryHandler.HandleCurrent(c.UserContext(), query.GetCurrentAssignment{
		CompanyID: companyID,
		UserID:    userID,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromAssignmentView(assignment))
}
