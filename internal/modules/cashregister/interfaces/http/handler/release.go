package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/command"
	"github.com/sigif/sigif-go/internal/modules/cashregister/application/dto"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// Release frees a cash register at the end of the shift (HU-10-02). The
// assigned cashier releases its own assignment; a user with
// cash_registers.release_any may release anyone's.
func (h *CashRegisterHTTPHandler) Release(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, errInvalidAssignmentID)
	}

	userID, companyID, err := actorFromContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	canReleaseAny, err := h.checker.HasPermission(c.UserContext(), userID, "cash_registers", "release_any")
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	assignment, err := h.cmdHandler.HandleRelease(c.UserContext(), command.ReleaseAssignment{
		CompanyID:     companyID,
		ActorID:       userID,
		AssignmentID:  id,
		CanReleaseAny: canReleaseAny,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromAssignmentView(assignment))
}
