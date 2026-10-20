package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashsession/application/command"
	"github.com/sigif/sigif-go/internal/modules/cashsession/application/dto"
	"github.com/sigif/sigif-go/internal/modules/cashsession/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// Open starts the cash session of the authenticated cashier on the cash
// register assigned to them (HU-009).
func (h *CashSessionHTTPHandler) Open(c *fiber.Ctx) error {
	var req dtos.OpenCashSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, errors.New(errors.CodeBadRequest, "invalid request body", fiber.StatusBadRequest))
	}
	req.Normalize()
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	userID, companyID, err := actorFromContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.OpenCashSession{
		CompanyID:     companyID,
		UserID:        userID,
		OpeningAmount: *req.OpeningAmount,
		OpeningNotes:  req.OpeningNotes,
	}
	if req.CashRegisterID != nil && *req.CashRegisterID != "" {
		id := uuid.MustParse(*req.CashRegisterID)
		cmd.CashRegisterID = &id
	}

	session, err := h.cmdHandler.HandleOpen(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Created(c, dto.FromSessionView(session))
}
