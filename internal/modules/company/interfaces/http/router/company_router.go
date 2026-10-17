package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/company/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func RegisterCompanyRoutes(router fiber.Router, h *handler.CompanyHTTPHandler, checker middleware.PermissionChecker) {
	router.Get("/companies", middleware.RequireAnyPermission(checker, "users.create", "users.update"), h.List)
}
