package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/logger"
	"github.com/sigif/sigif-go/internal/shared/response"
)

const PermissionDeniedMessage = "you do not have permission to perform this action"

type permissionCode struct {
	code      string
	module    string
	operation string
}

func parsePermissionCode(code string) permissionCode {
	module, operation, found := strings.Cut(code, ".")
	if !found || module == "" || operation == "" {
		panic(fmt.Sprintf("middleware: invalid permission code %q, expected module.operation", code))
	}
	return permissionCode{code: code, module: module, operation: operation}
}

func RequireAnyPermission(checker PermissionChecker, codes ...string) fiber.Handler {
	if len(codes) == 0 {
		panic("middleware: RequireAnyPermission needs at least one permission code")
	}
	required := make([]permissionCode, len(codes))
	for i, code := range codes {
		required[i] = parsePermissionCode(code)
	}

	return func(c *fiber.Ctx) error {
		userID, ok := UserIDFromContext(c.UserContext())
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}

		for _, permission := range required {
			allowed, err := checker.HasPermission(c.UserContext(), userID, permission.module, permission.operation)
			if err != nil {
				return response.Error(c, fiber.StatusInternalServerError, err)
			}
			if allowed {
				return c.Next()
			}
		}

		logPermissionDenied(c, userID, codes)
		return PermissionDenied(c)
	}
}

var deniedLogger = logger.L

func logPermissionDenied(c *fiber.Ctx, userID uuid.UUID, codes []string) {
	fields := []zap.Field{
		zap.String("user_id", userID.String()),
		zap.Strings("required_permissions", codes),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("ip", c.IP()),
		zap.String("request_id", c.GetRespHeader("X-Request-ID")),
	}
	if companyID, ok := CompanyIDFromContext(c.UserContext()); ok {
		fields = append(fields, zap.String("company_id", companyID.String()))
	}
	deniedLogger().Warn("permission denied", fields...)
}

func PermissionDenied(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusForbidden, errors.New(errors.CodeForbidden, PermissionDeniedMessage, fiber.StatusForbidden))
}
