package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

const contextKeyCashSessionID = "cash_session_id"

// CashSessionChecker resolves the open cash session a cashier needs to run a
// cash transaction. Implemented by the cash session module (RF-09).
type CashSessionChecker interface {
	RequireOpenSession(ctx context.Context, companyID, userID uuid.UUID) (uuid.UUID, error)
}

func WithCashSessionID(ctx context.Context, sessionID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyCashSessionID, sessionID)
}

func CashSessionIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	sessionID, ok := ctx.Value(contextKeyCashSessionID).(uuid.UUID)
	if !ok || sessionID == uuid.Nil {
		return uuid.Nil, false
	}
	return sessionID, true
}

// RequireOpenCashSession blocks a cash transaction when the cashier has no
// open cash session (HU-009, criteria 5 and 6) and answers
// 409 CASH_SESSION_REQUIRED. Place it after RequirePermission on every route
// that moves money (sales, payments, cash movements). The handler reads the
// session with CashSessionIDFromContext.
func RequireOpenCashSession(checker CashSessionChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := UserIDFromContext(c.UserContext())
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, errors.ErrUnauthorized)
		}
		companyID, ok := CompanyIDFromContext(c.UserContext())
		if !ok {
			return response.Error(c, fiber.StatusBadRequest, errors.New(errors.CodeBadRequest, "company is required", fiber.StatusBadRequest))
		}

		sessionID, err := checker.RequireOpenSession(c.UserContext(), companyID, userID)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, err)
		}

		c.SetUserContext(WithCashSessionID(c.UserContext(), sessionID))
		c.Locals(contextKeyCashSessionID, sessionID)
		return c.Next()
	}
}
