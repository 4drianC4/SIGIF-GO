package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/company/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func RegisterCompanyRoutes(router fiber.Router, h *handler.CompanyHTTPHandler, checker middleware.PermissionChecker) {
	router.Get("/companies", requireUserManagement(checker), h.List)
}

// The selector is available to administrators who can create OR edit users.
func requireUserManagement(checker middleware.PermissionChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, ok := middleware.UserIDFromContext(c.UserContext())
		if !ok {
			return response.Error(c, 401, errors.ErrUnauthorized)
		}
		for _, operation := range []string{"create", "update"} {
			allowed, err := checker.HasPermission(c.UserContext(), id, "users", operation)
			if err != nil {
				return response.Error(c, 500, err)
			}
			if allowed {
				return c.Next()
			}
		}
		return response.Error(c, 403, errors.ErrForbidden)
	}
}
