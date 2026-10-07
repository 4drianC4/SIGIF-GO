package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

// RegisterRoleRoutes exposes the role listing (HU-082-02). Roles are
// read-only: creation and assignment belong to other HUs.
func RegisterRoleRoutes(router fiber.Router, h *handler.RoleHTTPHandler, checker middleware.PermissionChecker) {
	roles := router.Group("/roles")

	roles.Get("/", middleware.RequirePermission(checker, "roles", "list"), h.List)
}
