package dtos

// AvailabilityAll lists cash registers in every availability.
const AvailabilityAll = "all"

// ListCashRegistersRequest holds the query string of GET /cash-registers. Page
// and limit are pointers so an explicit 0 is rejected instead of defaulted.
type ListCashRegistersRequest struct {
	Availability string `json:"availability" query:"availability" validate:"omitempty,oneof=available in_use all"`
	Page         *int   `json:"page" query:"page" validate:"omitempty,min=1"`
	Limit        *int   `json:"limit" query:"limit" validate:"omitempty,min=1"`
}

type AssignCashRegisterRequest struct {
	CashRegisterID string `json:"cash_register_id" validate:"required,uuid"`
}
