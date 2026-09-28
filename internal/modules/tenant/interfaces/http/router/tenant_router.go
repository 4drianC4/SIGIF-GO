package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/tenant/interfaces/http/handler"
)

func RegisterTenantRoutes(router fiber.Router, h *handler.TenantHTTPHandler) {
	tenants := router.Group("/tenants")
	tenants.Post("/", h.Create)
	tenants.Get("/", h.List)
	tenants.Get("/slug/:slug", h.GetBySlug)
	tenants.Get("/:id", h.GetByID)
	tenants.Put("/:id", h.Update)
	tenants.Delete("/:id", h.Delete)
	tenants.Post("/:id/activate", h.Activate)
	tenants.Post("/:id/deactivate", h.Deactivate)
}