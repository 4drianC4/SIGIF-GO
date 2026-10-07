package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

// RegisterPermissionRoutes exposes the permission catalog (HU-082-01). The
// static path /modules is registered before any dynamic :id route that later
// HUs may add under /permissions.
func RegisterPermissionRoutes(router fiber.Router, h *handler.PermissionHTTPHandler, checker middleware.PermissionChecker) {
	permissions := router.Group("/permissions")

	permissions.Get("/modules", middleware.RequirePermission(checker, "permissions", "read"), h.ListModules)
	permissions.Get("/", middleware.RequirePermission(checker, "permissions", "read"), h.List)
	permissions.Post("/", middleware.RequirePermission(checker, "permissions", "create"), h.Create)
}
