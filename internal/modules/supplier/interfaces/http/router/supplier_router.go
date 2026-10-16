package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func RegisterSupplierRoutes(router fiber.Router, h *handler.SupplierHTTPHandler, checker middleware.PermissionChecker) {
	suppliers := router.Group("/suppliers")
	suppliers.Get("/", middleware.RequirePermission(checker, "suppliers", "list"), h.List)
	suppliers.Get("/summary", middleware.RequirePermission(checker, "suppliers", "list"), h.Summary)
	suppliers.Get("/form-options", middleware.RequirePermission(checker, "suppliers", "list"), h.FormOptions)
	suppliers.Post("/", middleware.RequirePermission(checker, "suppliers", "create"), h.Create)
	suppliers.Get("/:id", middleware.RequirePermission(checker, "suppliers", "read"), h.Get)
	suppliers.Put("/:id", middleware.RequirePermission(checker, "suppliers", "update"), h.Update)
	suppliers.Patch("/:id/deactivate", middleware.RequirePermission(checker, "suppliers", "deactivate"), h.Deactivate)
	suppliers.Get("/:id/products", middleware.RequirePermission(checker, "suppliers", "read"), h.ListProducts)
	suppliers.Get("/:id/product-options", middleware.RequirePermission(checker, "suppliers", "update"), h.ProductOptions)
	suppliers.Post("/:id/products/batch", middleware.RequirePermission(checker, "suppliers", "update"), h.AssociateProducts)
	suppliers.Post("/:id/products", middleware.RequirePermission(checker, "suppliers", "update"), h.AssociateProduct)
	suppliers.Put("/:id/products/:product_id", middleware.RequirePermission(checker, "suppliers", "update"), h.UpdateAssociation)
	suppliers.Delete("/:id/products/:product_id", middleware.RequirePermission(checker, "suppliers", "update"), h.UnlinkProduct)

	// Both views use the same supplier-owned association implementation.
	products := router.Group("/products/:product_id")
	products.Get("/suppliers", middleware.RequirePermission(checker, "suppliers", "read"), h.ListProductSuppliers)
	products.Get("/supplier-options", middleware.RequirePermission(checker, "suppliers", "update"), h.SupplierOptions)
}
