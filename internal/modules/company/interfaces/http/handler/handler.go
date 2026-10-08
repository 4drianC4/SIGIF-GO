package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/company/application/handler"
	"github.com/sigif/sigif-go/internal/shared/response"
)

type CompanyHTTPHandler struct{ query *handler.CompanyQueryHandler }

func NewCompanyHTTPHandler(q *handler.CompanyQueryHandler) *CompanyHTTPHandler {
	return &CompanyHTTPHandler{query: q}
}
func (h *CompanyHTTPHandler) List(c *fiber.Ctx) error {
	options, err := h.query.HandleList(c.UserContext())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Success(c, options)
}
