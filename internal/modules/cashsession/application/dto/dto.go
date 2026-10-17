package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
)

// Money renders an amount as a JSON number with two decimals, as the catalog
// does with prices.
type Money decimal.Decimal

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(decimal.Decimal(m).StringFixed(2)), nil
}

type CashRegister struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Session is a cash session with its cash register.
type Session struct {
	ID            string        `json:"id"`
	Status        entity.Status `json:"status"`
	OpeningAmount Money         `json:"opening_amount"`
	OpeningNotes  *string       `json:"opening_notes"`
	OpenedAt      string        `json:"opened_at"`
	ClosedAt      *string       `json:"closed_at"`
	CashierID     string        `json:"cashier_id"`
	AssignmentID  string        `json:"assignment_id"`
	CashRegister  CashRegister  `json:"cash_register"`
}

// formatTime renders every timestamp in UTC so writes and reads look alike.
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func FromSessionView(v *service.SessionView) Session {
	s := v.Session
	result := Session{
		ID:            s.ID.String(),
		Status:        s.Status,
		OpeningAmount: Money(s.OpeningAmount),
		OpenedAt:      formatTime(s.OpenedAt),
		CashierID:     s.UserID.String(),
		AssignmentID:  s.AssignmentID.String(),
		CashRegister: CashRegister{
			ID:     v.Register.ID.String(),
			Code:   v.Register.Code,
			Name:   v.Register.Name,
			Status: v.Register.Status,
		},
	}
	if s.OpeningNotes != "" {
		notes := s.OpeningNotes
		result.OpeningNotes = &notes
	}
	if s.ClosedAt != nil {
		closedAt := formatTime(*s.ClosedAt)
		result.ClosedAt = &closedAt
	}
	return result
}
