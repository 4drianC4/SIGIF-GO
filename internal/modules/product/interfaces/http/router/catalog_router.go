package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func RegisterCatalogRoutes(router fiber.Router, h *handler.CatalogHTTPHandler, checker middleware.PermissionChecker) {
	router.Post("/categories", middleware.RequirePermission(checker, "categories", "create"), h.CreateCategory)

	router.Get("/products", middleware.RequirePermission(checker, "products", "read"), h.ListProducts)
	router.Post("/products", middleware.RequirePermission(checker, "products", "create"), h.CreateProduct)
	router.Get("/products/:id", middleware.RequirePermission(checker, "products", "read"), h.GetProduct)
	router.Put("/products/:id", middleware.RequirePermission(checker, "products", "update"), h.UpdateProduct)
	router.Patch("/products/:id/status", middleware.RequirePermission(checker, "products", "update"), h.SetProductStatus)
	router.Delete("/products/:id", middleware.RequirePermission(checker, "products", "delete"), h.DeleteProduct)
}
