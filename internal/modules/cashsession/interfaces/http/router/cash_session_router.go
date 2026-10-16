package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/cashsession/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func RegisterCashSessionRoutes(router fiber.Router, h *handler.CashSessionHTTPHandler, checker middleware.PermissionChecker) {
	sessions := router.Group("/cash-sessions")
	sessions.Post("/", middleware.RequirePermission(checker, "cash_sessions", "open"), h.Open)
	sessions.Get("/current", middleware.RequirePermission(checker, "cash_sessions", "read"), h.Current)
}
