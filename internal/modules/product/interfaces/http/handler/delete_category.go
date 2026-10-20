package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// DeleteCategory deletes a category logically. It answers 409 while the
// category still has subcategories or products.
func (h *CatalogHTTPHandler) DeleteCategory(c *fiber.Ctx) error {
	categoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, errInvalidCategoryID)
	}

	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, service.ErrCompanyRequired)
	}

	if err := h.cmdHandler.HandleDeleteCategory(c.UserContext(), command.DeleteCategory{
		CompanyID:  companyID,
		CategoryID: categoryID,
	}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.NoContent(c)
}
