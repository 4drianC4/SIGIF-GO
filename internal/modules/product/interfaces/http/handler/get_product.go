package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/application/dto"
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *CatalogHTTPHandler) GetProduct(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid product id", 400))
	}

	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	product, err := h.cmdHandler.HandleGetProduct(c.UserContext(), command.GetProduct{
		CompanyID: companyID,
		ProductID: productID,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromProduct(product))
}

func (h *CatalogHTTPHandler) ListProducts(c *fiber.Ctx) error {
	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	page, limit := pagination.Parse(c, 20, 100)

	cmd := command.ListProducts{
		CompanyID: companyID,
		Name:      c.Query("name"),
		Page:      page,
		Limit:     limit,
	}

	if catStr := c.Query("category_id"); catStr != "" {
		catID, err := uuid.Parse(catStr)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid category_id", 400))
		}
		cmd.CategoryID = &catID
	}

	if statusStr := c.Query("status"); statusStr != "" {
		s := entity.Status(statusStr)
		if s != entity.StatusActive && s != entity.StatusInactive {
			return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "status must be 'active' or 'inactive'", 400))
		}
		cmd.Status = &s
	}

	products, total, err := h.cmdHandler.HandleListProducts(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	page2, _ := pagination.Parse(c, 20, 100)
	meta := pagination.New(page2, limit, total)
	return response.Paginated(c, dto.FromProductList(products), meta)
}

func (h *CatalogHTTPHandler) SetProductStatus(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid product id", 400))
	}

	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}

	s := entity.Status(body.Status)
	if s != entity.StatusActive && s != entity.StatusInactive {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "status must be 'active' or 'inactive'", 400))
	}

	product, err := h.cmdHandler.HandleSetProductStatus(c.UserContext(), command.SetProductStatus{
		CompanyID: companyID,
		ProductID: productID,
		Status:    s,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromProduct(product))
}

func (h *CatalogHTTPHandler) DeleteProduct(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid product id", 400))
	}

	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	if err := h.cmdHandler.HandleDeleteProduct(c.UserContext(), command.DeleteProduct{
		CompanyID: companyID,
		ProductID: productID,
	}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}
