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

func (h *CatalogHTTPHandler) UpdateProductPrice(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, errInvalidProductID)
	}
	var req dtos.UpdateProductPriceRequest
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
	userID, ok := middleware.UserIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, sharedErrors.ErrUnauthorized)
	}
	product, err := h.cmdHandler.HandleUpdateProductPrice(c.UserContext(), command.UpdateProductPrice{
		CompanyID: companyID,
		ProductID: productID,
		UserID:    userID,
		Cost:      req.Cost,
		SalePrice: req.SalePrice,
		Reason:    req.Reason,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Success(c, dto.FromProduct(*product))
}
