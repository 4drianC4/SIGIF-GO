package entity

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

// MaxNotesLength is the longest opening note, in characters.
const MaxNotesLength = 200

// MaxOpeningAmount keeps the opening amount within numeric(12,2).
var MaxOpeningAmount = decimal.RequireFromString("9999999999.99")

// CashSession is the money shift of a cashier on the cash register assigned to
// them (RF-09). It starts open with a known cash amount; closing it belongs to
// RF-17.
type CashSession struct {
	ID             uuid.UUID
	CompanyID      uuid.UUID
	CashRegisterID uuid.UUID
	AssignmentID   uuid.UUID
	UserID         uuid.UUID
	OpeningAmount  decimal.Decimal
	OpeningNotes   string
	Status         Status
	OpenedAt       time.Time
	ClosedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

type NewSessionParams struct {
	CompanyID      uuid.UUID
	CashRegisterID uuid.UUID
	AssignmentID   uuid.UUID
	UserID         uuid.UUID
	OpeningAmount  decimal.Decimal
	OpeningNotes   string
}

// NewSession opens a session now. The amount and the notes are validated with
// ValidateOpening.
func NewSession(clk clock.Clock, params NewSessionParams) (*CashSession, error) {
	notes, err := ValidateOpening(params.OpeningAmount, params.OpeningNotes)
	if err != nil {
		return nil, err
	}
	now := clk.NowUTC()
	return &CashSession{
		ID:             uuid.New(),
		CompanyID:      params.CompanyID,
		CashRegisterID: params.CashRegisterID,
		AssignmentID:   params.AssignmentID,
		UserID:         params.UserID,
		OpeningAmount:  params.OpeningAmount,
		OpeningNotes:   notes,
		Status:         StatusOpen,
		OpenedAt:       now,
		CreatedAt:      now,
	}, nil
}

// ValidateOpening checks the opening amount (zero or more, at most two
// decimals) and the optional notes, and returns the notes trimmed.
func ValidateOpening(amount decimal.Decimal, notes string) (string, error) {
	details := map[string]string{}
	switch {
	case amount.IsNegative():
		details["opening_amount"] = "no puede ser negativo"
	case !amount.Equal(amount.Round(2)):
		details["opening_amount"] = "admite como máximo 2 decimales"
	case amount.GreaterThan(MaxOpeningAmount):
		details["opening_amount"] = "no puede superar " + MaxOpeningAmount.StringFixed(2)
	}
	notes = strings.TrimSpace(notes)
	if utf8.RuneCountInString(notes) > MaxNotesLength {
		details["opening_notes"] = "admite como máximo 200 caracteres"
	}
	if len(details) > 0 {
		return "", sharedErrors.New(sharedErrors.CodeValidation, "validation failed", 400).WithDetails(details)
	}
	return notes, nil
}

func (s *CashSession) IsOpen() bool {
	return s.Status == StatusOpen
}
