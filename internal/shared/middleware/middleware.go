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
	contextKeyUserID    = "user_id"
	contextKeyCompanyID = "company_id"
	contextKeyEmail     = "email"
	contextKeyRole      = "role"
	contextKeySessionID = "session_id"
)

// TokenHashLocalKey is the Fiber locals key where the authenticated request's
// token hash is stored for handlers that need it (e.g. logout).
const TokenHashLocalKey = "token_hash"

// SessionReader checks whether a stored session (identified by its token hash)
// is still active. Implemented by the auth module.
type SessionReader interface {
	IsActive(ctx context.Context, tokenHash string) (bool, error)
}

// PermissionChecker resolves whether a user holds a given module/operation
// permission. Implemented by the user module (RBAC).
type PermissionChecker interface {
	HasPermission(ctx context.Context, userID uuid.UUID, module, operation string) (bool, error)
}

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

func WithCompanyID(ctx context.Context, companyID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyCompanyID, companyID)
}

func WithSessionID(ctx context.Context, sessionID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeySessionID, sessionID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	value := ctx.Value(contextKeyUserID)
	if value == nil {
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, false
	}
	return userID, true
}

func CompanyIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	value := ctx.Value(contextKeyCompanyID)
	if value == nil {
		return uuid.Nil, false
	}
	companyID, ok := value.(uuid.UUID)
	if !ok || companyID == uuid.Nil {
		return uuid.Nil, false
	}
	return companyID, true
}

func SessionIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	value := ctx.Value(contextKeySessionID)
	if value == nil {
		return uuid.Nil, false
	}
	sessionID, ok := value.(uuid.UUID)
	if !ok || sessionID == uuid.Nil {
		return uuid.Nil, false
	}
	return sessionID, true
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
			if id, ok := c.Locals("requestid").(string); ok {
				requestID = id
			}
		}
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)
		return c.Next()
	}
}

func isPublicPath(path string) bool {
	return path == "/health" || path == "/api/v1/auth/login"
}

// AuthRequired authenticates the request from the Bearer token, verifies the
// corresponding session is active and stores the identity in the context.
func AuthRequired(jwtManager *jwt.JWTManager, sessions SessionReader) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if isPublicPath(c.Path()) {
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
		rawToken := strings.TrimSpace(parts[1])

		claims, err := jwtManager.ValidateAccessToken(rawToken)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}

		tokenHash := jwt.HashToken(rawToken)
		active, err := sessions.IsActive(c.UserContext(), tokenHash)
		if err != nil || !active {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}
		sessionID, err := uuid.Parse(claims.SessionID)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}

		ctx := c.UserContext()
		ctx = WithUserID(ctx, userID)
		ctx = WithSessionID(ctx, sessionID)
		if companyID, err := uuid.Parse(claims.CompanyID); err == nil && companyID != uuid.Nil {
			ctx = WithCompanyID(ctx, companyID)
		}
		if claims.Email != "" {
			ctx = context.WithValue(ctx, contextKeyEmail, claims.Email)
		}
		if claims.Role != "" {
			ctx = context.WithValue(ctx, contextKeyRole, claims.Role)
		}
		c.SetUserContext(ctx)

		c.Locals("user_id", userID)
		c.Locals("session_id", sessionID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals(TokenHashLocalKey, tokenHash)

		return c.Next()
	}
}

// RequirePermission enforces role-based access control for a module/operation.
func RequirePermission(checker PermissionChecker, module, operation string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := UserIDFromContext(c.UserContext())
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}

		allowed, err := checker.HasPermission(c.UserContext(), userID, module, operation)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, err)
		}
		if !allowed {
			return response.Error(c, fiber.StatusForbidden, errors.ErrForbidden)
		}

		return c.Next()
	}
}
