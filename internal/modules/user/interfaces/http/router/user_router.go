package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func RegisterUserRoutes(router fiber.Router, h *handler.UserHTTPHandler, checker middleware.PermissionChecker) {
	users := router.Group("/users")

	users.Post("/", middleware.RequirePermission(checker, "users", "create"), h.Create)
	users.Get("/", middleware.RequirePermission(checker, "users", "list"), h.List)
	users.Get("/by-email", middleware.RequirePermission(checker, "users", "read"), h.GetByEmail)
	users.Get("/:id", middleware.RequirePermission(checker, "users", "read"), h.GetByID)
	users.Patch("/:id", middleware.RequirePermission(checker, "users", "update"), h.Edit)
	users.Put("/:id/password", middleware.RequirePermission(checker, "users", "change_password"), h.ChangePassword)
	users.Delete("/:id", middleware.RequirePermission(checker, "users", "delete"), h.Delete)
	users.Post("/:id/activate", middleware.RequirePermission(checker, "users", "activate"), h.Activate)
	users.Post("/:id/deactivate", middleware.RequirePermission(checker, "users", "deactivate"), h.Deactivate)
}
