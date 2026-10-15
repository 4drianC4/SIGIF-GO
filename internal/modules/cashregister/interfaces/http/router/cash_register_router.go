package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/cashregister/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func RegisterCashRegisterRoutes(router fiber.Router, h *handler.CashRegisterHTTPHandler, checker middleware.PermissionChecker) {
	registers := router.Group("/cash-registers")
	registers.Get("/", middleware.RequirePermission(checker, "cash_registers", "list"), h.ListCashRegisters)

	assignments := router.Group("/cash-register-assignments")
	assignments.Post("/", middleware.RequirePermission(checker, "cash_registers", "assign"), h.Assign)
	// Static path before the :id routes so they never collide.
	assignments.Get("/current", middleware.RequirePermission(checker, "cash_registers", "assign"), h.Current)
	assignments.Post("/:id/release", requireAnyPermission(checker, "cash_registers", "release", "release_any"), h.Release)
}

// requireAnyPermission lets the request through when the user holds at least
// one of the operations of the module.
func requireAnyPermission(checker middleware.PermissionChecker, module string, operations ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := middleware.UserIDFromContext(c.UserContext())
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}
		for _, operation := range operations {
			allowed, err := checker.HasPermission(c.UserContext(), userID, module, operation)
			if err != nil {
				return response.Error(c, fiber.StatusInternalServerError, err)
			}
			if allowed {
				return c.Next()
			}
		}
		return response.Error(c, fiber.StatusForbidden, errors.ErrForbidden)
	}
}
