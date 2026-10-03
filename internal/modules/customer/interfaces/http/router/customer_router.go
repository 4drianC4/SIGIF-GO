package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

// RegisterCustomerRoutes registra todas las rutas REST para clientes.
func RegisterCustomerRoutes(router fiber.Router, h *handler.CustomerHTTPHandler, jwtManager *jwt.JWTManager) {
	// Middleware de autenticación y tenant
	authMiddleware := middleware.RequireAuth(jwtManager)
	tenantMiddleware := middleware.TenantContext()

	customers := router.Group("/customers", authMiddleware, tenantMiddleware)
	customers.Post("/", h.Create)
	customers.Get("/", h.List)
	customers.Get("/:id", h.GetByID)
	customers.Put("/:id", h.Update)
	customers.Patch("/:id/status", h.ChangeStatus)
	customers.Delete("/:id", h.Delete)
}
