package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

// sessionCheckerAdapter exposes the cash session rule to the shared
// RequireOpenCashSession middleware, so other modules can guard their cash
// transactions without depending on this module.
type sessionCheckerAdapter struct {
	sessions *service.CashSessionService
}

func NewSessionCheckerAdapter(sessions *service.CashSessionService) middleware.CashSessionChecker {
	return &sessionCheckerAdapter{sessions: sessions}
}

func (a *sessionCheckerAdapter) RequireOpenSession(ctx context.Context, companyID, userID uuid.UUID) (uuid.UUID, error) {
	session, err := a.sessions.RequireOpenSession(ctx, companyID, userID)
	if err != nil {
		return uuid.Nil, err
	}
	return session.ID, nil
}
