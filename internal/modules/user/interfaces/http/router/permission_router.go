package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

// RegisterPermissionRoutes exposes the permission catalog (HU-082-01). The
// static paths /modules and /export are registered before the dynamic :id
// routes so they are never captured as a permission id.
func RegisterPermissionRoutes(router fiber.Router, h *handler.PermissionHTTPHandler, checker middleware.PermissionChecker) {
	permissions := router.Group("/permissions")

	permissions.Get("/modules", middleware.RequirePermission(checker, "permissions", "read"), h.ListModules)
	permissions.Get("/export", middleware.RequirePermission(checker, "permissions", "export"), h.Export)
	permissions.Get("/", middleware.RequirePermission(checker, "permissions", "read"), h.List)
	permissions.Post("/", middleware.RequirePermission(checker, "permissions", "create"), h.Create)
	permissions.Patch("/:id", middleware.RequirePermission(checker, "permissions", "update"), h.Update)
	permissions.Delete("/:id", middleware.RequirePermission(checker, "permissions", "delete"), h.Delete)
}
