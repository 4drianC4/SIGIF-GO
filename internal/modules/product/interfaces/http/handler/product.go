package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/application/dto"
	"github.com/sigif/sigif-go/internal/modules/product/application/query"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/pagination"
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

	created, err := h.cmdHandler.HandleCreateProduct(c.UserContext(), command.CreateProduct{
		CompanyID:       companyID,
		CategoryID:      uuid.MustParse(req.CategoryID),
		UnitOfMeasureID: uuid.MustParse(req.UnitOfMeasureID),
		SKU:             req.SKU,
		Barcode:         req.Barcode,
		Name:            req.Name,
		Description:     req.Description,
		Cost:            *req.Cost,
		SalePrice:       *req.SalePrice,
		InitialStock:    dtos.OrZero(req.InitialStock),
		MinStock:        dtos.OrZero(req.MinStock),
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Created(c, dto.FromProduct(*created))
}

func (h *CatalogHTTPHandler) ListProducts(c *fiber.Ctx) error {
	var q dtos.ListProductsQuery
	if err := c.QueryParser(&q); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	if err := h.validator.Validate(&q); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	page, limit := pagination.Parse(c, 20, 100)
	products, total, err := h.queryHandler.HandleListProducts(c.UserContext(), query.ListProducts{
		CompanyID:  companyID,
		Search:     q.Search,
		CategoryID: q.CategoryUUID(),
		Status:     q.StatusFilter(),
		Offset:     (page - 1) * limit,
		Limit:      limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Paginated(c, dto.FromProductList(products), pagination.New(page, limit, total))
}

func (h *CatalogHTTPHandler) ProductSummary(c *fiber.Ctx) error {
	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	summary, err := h.queryHandler.HandleProductSummary(c.UserContext(), companyID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromSummary(summary))
}

func (h *CatalogHTTPHandler) ValidateDuplicate(c *fiber.Ctx) error {
	var q dtos.ValidateDuplicateQuery
	if err := c.QueryParser(&q); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	if err := h.validator.Validate(&q); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	result, err := h.queryHandler.HandleValidateDuplicate(c.UserContext(), query.ValidateDuplicate{
		CompanyID: companyID,
		Name:      q.Name,
		SKU:       q.SKU,
		Barcode:   strings.TrimSpace(q.Barcode),
		ExcludeID: q.ExcludeUUID(),
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromDuplicateResult(result))
}
