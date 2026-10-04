package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func RegisterCatalogRoutes(router fiber.Router, h *handler.CatalogHTTPHandler, checker middleware.PermissionChecker) {
	router.Post("/categories", middleware.RequirePermission(checker, "categories", "create"), h.CreateCategory)
	router.Post("/products", middleware.RequirePermission(checker, "products", "create"), h.CreateProduct)
}
