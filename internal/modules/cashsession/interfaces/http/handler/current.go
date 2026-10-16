package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/cashsession/application/dto"
	"github.com/sigif/sigif-go/internal/modules/cashsession/application/query"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// Current returns the open cash session of the authenticated cashier, or
// 404 NO_OPEN_CASH_SESSION when there is none (HU-009).
func (h *CashSessionHTTPHandler) Current(c *fiber.Ctx) error {
	userID, companyID, err := actorFromContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	session, err := h.queryHandler.HandleCurrent(c.UserContext(), query.GetCurrentSession{
		CompanyID: companyID,
		UserID:    userID,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Success(c, dto.FromSessionView(session))
}
