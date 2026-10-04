package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func RegisterCustomerRoutes(router fiber.Router, h *handler.CustomerHTTPHandler, checker middleware.PermissionChecker) {
	customers := router.Group("/customers")
	customers.Post("/", middleware.RequirePermission(checker, "customers", "create"), h.Create)
	customers.Get("/", middleware.RequirePermission(checker, "customers", "list"), h.List)
	customers.Get("/:id", middleware.RequirePermission(checker, "customers", "read"), h.GetByID)
	customers.Put("/:id", middleware.RequirePermission(checker, "customers", "update"), h.Update)
	customers.Patch("/:id/status", middleware.RequirePermission(checker, "customers", "status"), h.ChangeStatus)
	customers.Delete("/:id", middleware.RequirePermission(checker, "customers", "delete"), h.Delete)
}
