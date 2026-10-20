package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/application/dto"
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
)

var errInvalidCategoryID = sharedErrors.New(sharedErrors.CodeBadRequest, "invalid category id", 400)

// UpdateCategory partially updates a category, parent or subcategory. The body
// is strict: it must carry at least one editable field and nothing else.
func (h *CatalogHTTPHandler) UpdateCategory(c *fiber.Ctx) error {
	categoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, errInvalidCategoryID)
	}

	req, cleared, details, err := dtos.ParseCategoryPatch(c.Body())
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := h.validator.Validate(&req); err != nil {
		var appErr *sharedErrors.AppError
		if !errors.As(err, &appErr) {
			return response.Error(c, fiber.StatusBadRequest, err)
		}
		fieldErrors, ok := appErr.Details.(map[string]string)
		if !ok {
			return response.Error(c, fiber.StatusBadRequest, err)
		}
		for field, message := range fieldErrors {
			details[field] = message
		}
	}
	if len(details) > 0 {
		return response.ValidationError(c, details)
	}

	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	cmd := command.UpdateCategory{CompanyID: companyID, CategoryID: categoryID, Name: req.Name, Description: req.Description}
	if cleared["description"] {
		empty := ""
		cmd.Description = &empty
	}
	switch {
	case req.DefaultTax != nil:
		cmd.DefaultTax = entity.Some(*req.DefaultTax)
	case cleared["default_tax"]:
		cmd.DefaultTax = entity.Cleared[string]()
	}
	switch {
	case req.TargetMargin != nil:
		cmd.TargetMargin = entity.Some(*req.TargetMargin)
	case cleared["target_margin"]:
		cmd.TargetMargin = entity.Cleared[decimal.Decimal]()
	}

	item, err := h.cmdHandler.HandleUpdateCategory(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Success(c, dto.FromCategory(item.Category, item.ProductsCount, item.DefaultTaxName))
}
