package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
)

func RegisterUserRoutes(router fiber.Router, h *handler.UserHTTPHandler) {
	users := router.Group("/users")
	users.Post("/", h.Create)
	users.Get("/", h.List)
	users.Get("/by-email", h.GetByEmail)
	users.Get("/:id", h.GetByID)
	users.Put("/:id", h.Update)
	users.Put("/:id/password", h.ChangePassword)
	users.Delete("/:id", h.Delete)
	users.Post("/:id/activate", h.Activate)
	users.Post("/:id/deactivate", h.Deactivate)
	users.Post("/:id/suspend", h.Suspend)
}