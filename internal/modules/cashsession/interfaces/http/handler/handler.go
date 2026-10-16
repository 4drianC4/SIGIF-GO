package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashsession/application/handler"
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type CashSessionHTTPHandler struct {
	cmdHandler   *handler.CashSessionCommandHandler
	queryHandler *handler.CashSessionQueryHandler
	validator    *validator.Validator
}

func NewCashSessionHTTPHandler(
	cmdHandler *handler.CashSessionCommandHandler,
	queryHandler *handler.CashSessionQueryHandler,
	val *validator.Validator,
) *CashSessionHTTPHandler {
	return &CashSessionHTTPHandler{cmdHandler: cmdHandler, queryHandler: queryHandler, validator: val}
}

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
