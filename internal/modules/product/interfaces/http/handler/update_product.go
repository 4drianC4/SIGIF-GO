package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/application/dto"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *CatalogHTTPHandler) UpdateProduct(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, errInvalidProductID)
	}

	var req dtos.UpdateProductRequest
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

	product, err := h.cmdHandler.HandleUpdateProduct(c.UserContext(), command.UpdateProduct{
		CompanyID:       companyID,
		ProductID:       productID,
		CategoryID:      uuid.MustParse(req.CategoryID),
		UnitOfMeasureID: uuid.MustParse(req.UnitOfMeasureID),
		SKU:             req.SKU,
		Barcode:         req.Barcode,
		Name:            req.Name,
		Description:     req.Description,
		Cost:            *req.Cost,
		SalePrice:       *req.SalePrice,
		MinStock:        req.MinStock,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromProduct(*product))
}
