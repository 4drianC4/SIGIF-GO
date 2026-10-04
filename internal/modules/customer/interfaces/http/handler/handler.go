package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/application/handler"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type CustomerHTTPHandler struct {
	cmdHandler   *handler.CustomerCommandHandler
	queryHandler *handler.CustomerQueryHandler
	validator    *validator.Validator
	cfg          *config.Config
}

func NewCustomerHTTPHandler(
	cmdHandler *handler.CustomerCommandHandler,
	queryHandler *handler.CustomerQueryHandler,
	val *validator.Validator,
	cfg *config.Config,
) *CustomerHTTPHandler {
	return &CustomerHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
		validator:    val,
		cfg:          cfg,
	}
}

// companyIDFromContext returns the authenticated user's company, or a 400 error.
func companyIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return uuid.Nil, response.Error(c, fiber.StatusBadRequest,
			errors.New(errors.CodeBadRequest, "company is required", fiber.StatusBadRequest))
	}
	return companyID, nil
}
