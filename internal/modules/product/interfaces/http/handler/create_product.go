package handler

import (
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

func (h *CatalogHTTPHandler) CreateProduct(c *fiber.Ctx) error {
	var req dtos.CreateProductRequest
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

	categoryID := uuid.MustParse(req.CategoryID)

	costPrice := decimal.Zero
	if req.CostPrice != nil {
		costPrice = *req.CostPrice
	}

	product, err := h.cmdHandler.HandleCreateProduct(c.UserContext(), command.CreateProduct{
		CompanyID:     companyID,
		CategoryID:    categoryID,
		SKU:           req.SKU,
		Barcode:       req.Barcode,
		Name:          req.Name,
		Description:   req.Description,
		UnitOfMeasure: entity.UnitOfMeasure(req.UnitOfMeasure),
		CostPrice:     costPrice,
		SalePrice:     *req.SalePrice,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Created(c, dto.FromProduct(product))
}
