package dtos

import (
	"strings"

	"github.com/shopspring/decimal"
)

// OpenCashSessionRequest is the body of POST /cash-sessions. The cashier and
// the company come from the token. CashRegisterID is optional: the session
// opens on the cash register assigned to the cashier, and when it is sent it
// must be that one.
type OpenCashSessionRequest struct {
	CashRegisterID *string          `json:"cash_register_id" validate:"omitempty,uuid"`
	OpeningAmount  *decimal.Decimal `json:"opening_amount" validate:"required"`
	OpeningNotes   string           `json:"opening_notes" validate:"omitempty,max=200"`
}

func (r *OpenCashSessionRequest) Normalize() {
	r.OpeningNotes = strings.TrimSpace(r.OpeningNotes)
	if r.CashRegisterID != nil {
		id := strings.TrimSpace(*r.CashRegisterID)
		r.CashRegisterID = &id
	}
}
