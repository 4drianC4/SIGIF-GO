package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/company/interfaces/http/handler"
)

func RegisterCompanyRoutes(router fiber.Router, h *handler.CompanyHTTPHandler) {
	companies := router.Group("/companies")
	companies.Post("/", h.Create)
	companies.Get("/", h.List)
	companies.Get("/:id", h.GetByID)
	companies.Put("/:id", h.Update)
	companies.Delete("/:id", h.Delete)
	companies.Post("/:id/activate", h.Activate)
	companies.Post("/:id/deactivate", h.Deactivate)
}