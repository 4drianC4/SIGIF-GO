package middleware

import (
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/logger"
	"github.com/sigif/sigif-go/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

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
		if tenantID != "" {
			c.Locals("tenant_id", tenantID)
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