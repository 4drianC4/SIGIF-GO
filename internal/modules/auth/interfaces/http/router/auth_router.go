package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/handler"
)

func RegisterAuthRoutes(router fiber.Router, h *handler.AuthHTTPHandler) {
	auth := router.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/logout", h.Logout)
	auth.Get("/me", h.Me)
	auth.Get("/me/permissions", h.MePermissions)
}
