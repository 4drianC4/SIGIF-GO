package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func RegisterCatalogRoutes(router fiber.Router, h *handler.CatalogHTTPHandler, checker middleware.PermissionChecker) {
	products := router.Group("/products")
	products.Get("/", middleware.RequirePermission(checker, "products", "list"), h.ListProducts)
	products.Get("/summary", middleware.RequirePermission(checker, "products", "list"), h.ProductSummary)
	products.Get("/validate-duplicate", middleware.RequirePermission(checker, "products", "create"), h.ValidateDuplicate)
	products.Post("/", middleware.RequirePermission(checker, "products", "create"), h.CreateProduct)

	categories := router.Group("/categories")
	categories.Get("/", middleware.RequirePermission(checker, "categories", "list"), h.ListCategories)
	categories.Post("/", middleware.RequirePermission(checker, "categories", "create"), h.CreateCategory)

	router.Get("/units-of-measure", h.ListUnitsOfMeasure)
	router.Get("/taxes", h.ListTaxes)
}
