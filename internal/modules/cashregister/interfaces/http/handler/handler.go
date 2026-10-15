package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/handler"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type CashRegisterHTTPHandler struct {
	cmdHandler   *handler.CashRegisterCommandHandler
	queryHandler *handler.CashRegisterQueryHandler
	checker      middleware.PermissionChecker
	validator    *validator.Validator
	cfg          *config.Config
}

func NewCashRegisterHTTPHandler(
	cmdHandler *handler.CashRegisterCommandHandler,
	queryHandler *handler.CashRegisterQueryHandler,
	checker middleware.PermissionChecker,
	val *validator.Validator,
	cfg *config.Config,
) *CashRegisterHTTPHandler {
	return &CashRegisterHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
		checker:      checker,
		validator:    val,
		cfg:          cfg,
	}
}

var errInvalidAssignmentID = errors.New(errors.CodeBadRequest, "invalid assignment id", fiber.StatusBadRequest)

// actorFromContext returns the authenticated user and its company. The
// cashier and the company always come from the token, never from the body.
func actorFromContext(c *fiber.Ctx) (userID, companyID uuid.UUID, err error) {
	userID, ok := middleware.UserIDFromContext(c.UserContext())
	if !ok {
		return uuid.Nil, uuid.Nil, errors.ErrUnauthorized
	}
	companyID, ok = middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return uuid.Nil, uuid.Nil, errors.New(errors.CodeBadRequest, "company is required", fiber.StatusBadRequest)
	}
	return userID, companyID, nil
}
