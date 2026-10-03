package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// RequireRoles deja pasar solo a usuarios con al menos uno de los roles indicados.
// Debe ir después de AuthRequired, que guarda los roles del token en c.Locals("roles").
func RequireRoles(allowed ...string) fiber.Handler {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, role := range allowed {
		allowedSet[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		roles, ok := c.Locals(contextKeyRoles).([]string)
		if !ok {
			if c.Locals(contextKeyUserID) == nil {
				return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
			}
			return response.Error(c, fiber.StatusForbidden, errors.ErrForbidden)
		}

		for _, role := range roles {
			if _, ok := allowedSet[role]; ok {
				return c.Next()
			}
		}
		return response.Error(c, fiber.StatusForbidden, errors.ErrForbidden)
	}
}
