package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/application/dto"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *CatalogHTTPHandler) CreateCategory(c *fiber.Ctx) error {
	var req dtos.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	req.Normalize()
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	created, err := h.cmdHandler.HandleCreateCategory(c.UserContext(), command.CreateCategory{
		CompanyID:    companyID,
		ParentID:     req.ParentUUID(),
		Name:         req.Name,
		Description:  req.Description,
		DefaultTax:   req.DefaultTax,
		TargetMargin: req.TargetMargin,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Created(c, dto.FromCreatedCategory(created))
}

func (h *CatalogHTTPHandler) ListCategories(c *fiber.Ctx) error {
	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	categories, err := h.queryHandler.HandleListCategories(c.UserContext(), companyID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromCategoryList(categories))
}
