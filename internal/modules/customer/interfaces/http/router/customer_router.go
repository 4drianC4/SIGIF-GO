package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/handler"
)

func RegisterCustomerRoutes(router fiber.Router, h *handler.CustomerHTTPHandler) {
    customers := router.Group("/customers")
    customers.Post("/", h.Create)
    customers.Get("/", h.List)
    customers.Get("/:id", h.GetByID)
    customers.Put("/:id", h.Update)
    customers.Patch("/:id/status", h.ChangeStatus)
    customers.Delete("/:id", h.Delete)
}