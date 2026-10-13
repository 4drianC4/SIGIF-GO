package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/product/application/dto"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *CatalogHTTPHandler) ListUnitsOfMeasure(c *fiber.Ctx) error {
	units, err := h.queryHandler.HandleListUnitsOfMeasure(c.UserContext())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Success(c, dto.FromUnits(units))
}

func (h *CatalogHTTPHandler) ListTaxes(c *fiber.Ctx) error {
	taxes, err := h.queryHandler.HandleListTaxes(c.UserContext())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Success(c, dto.FromTaxes(taxes))
}
