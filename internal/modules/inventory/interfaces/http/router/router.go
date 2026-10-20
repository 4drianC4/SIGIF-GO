package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/inventory/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func RegisterInventoryRoutes(router fiber.Router, h *handler.InventoryHTTPHandler, checker middleware.PermissionChecker) {
	inventory := router.Group("/inventory")
	inventory.Post("/movements", middleware.RequirePermission(checker, "inventory", "create"), h.RegisterMovement)
	inventory.Get("/movements", middleware.RequirePermission(checker, "inventory", "list"), h.ListMovements)
	inventory.Get("/branches", middleware.RequirePermission(checker, "inventory", "list"), h.ListBranches)
	inventory.Get("/stock", middleware.RequirePermission(checker, "inventory", "read"), h.ProductStock)
}
