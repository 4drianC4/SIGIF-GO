package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// RequireAuth valida el Bearer token JWT y puebla en contexto:
//   - "user_id"   → string
//   - "tenant_id" → uuid.UUID  (sustituye al valor crudo de X-Tenant-ID)
//   - "roles"     → []string
//
// Retorna 401 si el token falta, es inválido o expiró.
func RequireAuth(jwtManager *jwt.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized,
				sharedErrors.New(sharedErrors.CodeUnauthorized, "authorization header required", fiber.StatusUnauthorized))
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Error(c, fiber.StatusUnauthorized,
				sharedErrors.New(sharedErrors.CodeUnauthorized, "invalid authorization format", fiber.StatusUnauthorized))
		}

		claims, err := jwtManager.Validate(parts[1])
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized,
				sharedErrors.New(sharedErrors.CodeUnauthorized, "invalid or expired token", fiber.StatusUnauthorized))
		}

		// Parsear tenant_id del claim como UUID
		tenantID, err := uuid.Parse(claims.TenantID)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized,
				sharedErrors.New(sharedErrors.CodeUnauthorized, "invalid tenant in token", fiber.StatusUnauthorized))
		}

		// Publicar valores confiables en el contexto local de Fiber
		c.Locals("user_id", claims.UserID)
		c.Locals("tenant_id", tenantID)
		c.Locals("roles", claims.Roles)

		return c.Next()
	}
}
