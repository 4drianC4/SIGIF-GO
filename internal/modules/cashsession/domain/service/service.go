package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type CashSessionService struct {
	repo      repository.SessionRepository
	registers repository.RegisterAccess
	clock     clock.Clock
}

func NewCashSessionService(
	repo repository.SessionRepository,
	registers repository.RegisterAccess,
	clk clock.Clock,
) *CashSessionService {
	return &CashSessionService{repo: repo, registers: registers, clock: clk}
}

// SessionView is a cash session with its cash register.
type SessionView struct {
	Session  *entity.CashSession
	Register repository.RegisterInfo
}

type OpenParams struct {
	CompanyID uuid.UUID
	UserID    uuid.UUID
	// CashRegisterID is optional: the session always opens on the cash
	// register assigned to the cashier, and when it is given it must be
	// that one.
	CashRegisterID *uuid.UUID
	OpeningAmount  decimal.Decimal
	OpeningNotes   string
}

// Open starts the cash session of the cashier on the cash register they hold
// (HU-009). The cash register is never released here.
func (s *CashSessionService) Open(ctx context.Context, params OpenParams) (*SessionView, error) {
	if _, err := entity.ValidateOpening(params.OpeningAmount, params.OpeningNotes); err != nil {
		return nil, err
	}
	if params.CashRegisterID != nil {
		if err := s.registers.EnsureHolds(ctx, params.CompanyID, params.UserID, *params.CashRegisterID); err != nil {
			return nil, err
		}
	}
	hold, err := s.registers.ActiveHold(ctx, params.CompanyID, params.UserID)
	if err != nil {
		return nil, err
	}
	if !hold.Register.IsActive() {
		return nil, ErrCashRegisterInactive
	}

	current, err := s.repo.OpenByUser(ctx, params.CompanyID, params.UserID)
	if err != nil {
		return nil, err
	}
	if current != nil {
		return nil, ErrCashierHasOpenSession
	}
	other, err := s.repo.OpenByRegister(ctx, params.CompanyID, hold.Register.ID)
	if err != nil {
		return nil, err
	}
	if other != nil {
		return nil, ErrRegisterHasOpenSession
	}

	session, err := entity.NewSession(s.clock, entity.NewSessionParams{
		CompanyID:      params.CompanyID,
		CashRegisterID: hold.Register.ID,
		AssignmentID:   hold.AssignmentID,
		UserID:         params.UserID,
		OpeningAmount:  params.OpeningAmount,
		OpeningNotes:   params.OpeningNotes,
	})
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}
	return &SessionView{Session: session, Register: hold.Register}, nil
}

// Current returns the open session of the cashier, or ErrNoOpenSession.
func (s *CashSessionService) Current(ctx context.Context, companyID, userID uuid.UUID) (*SessionView, error) {
	session, err := s.repo.OpenByUser(ctx, companyID, userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrNoOpenSession
	}
	register, err := s.registers.Describe(ctx, companyID, session.CashRegisterID)
	if err != nil {
		return nil, err
	}
	return &SessionView{Session: session, Register: *register}, nil
}

// RequireOpenSession is the check every cash transaction runs first (HU-009,
// criteria 5 and 6): the cashier needs an open session on a cash register
// that is still assigned to them. Without a session it returns
// ErrSessionRequired.
func (s *CashSessionService) RequireOpenSession(ctx context.Context, companyID, userID uuid.UUID) (*entity.CashSession, error) {
	session, err := s.repo.OpenByUser(ctx, companyID, userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionRequired
	}
	if err := s.registers.EnsureHolds(ctx, companyID, userID, session.CashRegisterID); err != nil {
		return nil, err
	}
	return session, nil
}
