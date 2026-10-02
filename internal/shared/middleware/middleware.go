package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/logger"
	"github.com/sigif/sigif-go/internal/shared/response"
)

const (
	contextKeyTenantID = "tenant_id"
	contextKeyUserID   = "user_id"
	contextKeyEmail    = "email"
	contextKeyRoles    = "roles"
)

func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyTenantID, tenantID)
}

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	value := ctx.Value(contextKeyTenantID)
	if value == nil {
		return uuid.Nil, false
	}
	tenantID, ok := value.(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		return uuid.Nil, false
	}
	return tenantID, true
}

func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		var appErr *errors.AppError

		if e, ok := err.(*errors.AppError); ok {
			appErr = e
			code = e.StatusCode
		} else if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			appErr = errors.New(errors.CodeInternal, e.Message, code)
		} else {
			appErr = errors.New(errors.CodeInternal, err.Error(), code)
		}

		logger.L().Error("request error",
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.String("request_id", c.Get("X-Request-ID")),
			zap.Error(err),
		)

		return response.Error(c, code, appErr)
	}
}

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = c.Locals("requestid").(string)
		}
		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)
		return c.Next()
	}
}

func TenantContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Get("X-Tenant-ID")
		if tenantID == "" {
			return c.Next()
		}

		if parsed, err := uuid.Parse(tenantID); err == nil {
			c.Locals("tenant_id", parsed)
			c.SetUserContext(WithTenantID(c.UserContext(), parsed))
			return c.Next()
		}

		c.Locals("tenant_id", uuid.Nil)
		c.Locals("tenant_error", errors.ErrInvalidTenant)
		return c.Next()
	}
}

func AuthRequired(jwtManager *jwt.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if path == "/health" || path == "/auth/login" || path == "/auth/refresh" {
			return c.Next()
		}

		authorization := c.Get("Authorization")
		if authorization == "" {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}

		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}

		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}

		if claims.UserID == "" {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}

		if parsedUserID, err := uuid.Parse(claims.UserID); err == nil {
			c.Locals("user_id", parsedUserID)
			c.SetUserContext(WithUserID(c.UserContext(), parsedUserID))
		}
		if parsedTenantID, err := uuid.Parse(claims.TenantID); err == nil {
			c.Locals("tenant_id", parsedTenantID)
			c.SetUserContext(WithTenantID(c.UserContext(), parsedTenantID))
		}
		if claims.Email != "" {
			c.Locals("email", claims.Email)
			c.SetUserContext(context.WithValue(c.UserContext(), contextKeyEmail, claims.Email))
		}
		if len(claims.Roles) > 0 {
			c.Locals("roles", claims.Roles)
			c.SetUserContext(context.WithValue(c.UserContext(), contextKeyRoles, claims.Roles))
		}

		return c.Next()
	}
}

func Recovery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				logger.L().Error("panic recovered",
					zap.String("path", c.Path()),
					zap.Any("panic", r),
				)
				_ = response.Error(c, fiber.StatusInternalServerError, errors.ErrInternal)
			}
		}()
		return c.Next()
	}
}