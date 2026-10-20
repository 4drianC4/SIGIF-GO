package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/handler"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/service"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type SupplierHTTPHandler struct {
	commands  *handler.SupplierCommandHandler
	queries   *handler.SupplierQueryHandler
	validator *validator.Validator
}

func NewSupplierHTTPHandler(commands *handler.SupplierCommandHandler, queries *handler.SupplierQueryHandler, val *validator.Validator) *SupplierHTTPHandler {
	return &SupplierHTTPHandler{commands: commands, queries: queries, validator: val}
}

func company(c *fiber.Ctx) (uuid.UUID, error) {
	id, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return uuid.Nil, service.ErrCompanyRequired
	}
	return id, nil
}
func pathUUID(c *fiber.Ctx, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(key))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid "+key, 400)
	}
	return id, nil
}
func identity(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	companyID, err := company(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	id, err := pathUUID(c, "id")
	return companyID, id, err
}
func fail(c *fiber.Ctx, err error) error {
	return response.Error(c, fiber.StatusInternalServerError, err)
}

func (h *SupplierHTTPHandler) body(c *fiber.Ctx, dest any) error {
	if err := c.BodyParser(dest); err != nil {
		return sharedErrors.ErrBadRequest
	}
	if normalizer, ok := dest.(interface{ Normalize() }); ok {
		normalizer.Normalize()
	}
	return h.validator.Validate(dest)
}
func (h *SupplierHTTPHandler) query(c *fiber.Ctx, dest any) error {
	if err := c.QueryParser(dest); err != nil {
		return sharedErrors.ErrBadRequest
	}
	return h.validator.Validate(dest)
}
