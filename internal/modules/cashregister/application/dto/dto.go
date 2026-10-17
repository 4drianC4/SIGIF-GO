package dto

import (
	"time"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
)

// Cashier identifies the user that holds a cash register.
type Cashier struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CashRegisterItem is a cash register in a listing, with its availability.
type CashRegisterItem struct {
	ID           string                    `json:"id"`
	Code         string                    `json:"code"`
	Name         string                    `json:"name"`
	Status       entity.CashRegisterStatus `json:"status"`
	Availability entity.Availability       `json:"availability"`
	AssignmentID *string                   `json:"assignment_id,omitempty"`
	AssignedTo   *Cashier                  `json:"assigned_to,omitempty"`
	AssignedAt   *string                   `json:"assigned_at,omitempty"`
}

// CashRegister is the cash register embedded in an assignment.
type CashRegister struct {
	ID     string                    `json:"id"`
	Code   string                    `json:"code"`
	Name   string                    `json:"name"`
	Status entity.CashRegisterStatus `json:"status"`
}

// Assignment is a cash register assignment with its cash register and cashier.
type Assignment struct {
	ID           string                  `json:"id"`
	Status       entity.AssignmentStatus `json:"status"`
	CashRegister CashRegister            `json:"cash_register"`
	Cashier      Cashier                 `json:"cashier"`
	AssignedAt   string                  `json:"assigned_at"`
	ReleasedAt   *string                 `json:"released_at,omitempty"`
	ReleasedBy   *string                 `json:"released_by,omitempty"`
}

// formatTime renders every timestamp in UTC so writes and reads look alike.
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func FromRegisterViews(views []service.RegisterView) []CashRegisterItem {
	items := make([]CashRegisterItem, len(views))
	for i, v := range views {
		items[i] = CashRegisterItem{
			ID:           v.Register.ID.String(),
			Code:         v.Register.Code,
			Name:         v.Register.Name,
			Status:       v.Register.Status,
			Availability: entity.AvailabilityAvailable,
		}
		if v.Assignment != nil {
			assignmentID := v.Assignment.ID.String()
			assignedAt := formatTime(v.Assignment.AssignedAt)
			items[i].Availability = entity.AvailabilityInUse
			items[i].AssignmentID = &assignmentID
			items[i].AssignedTo = &Cashier{ID: v.Assignment.UserID.String(), Name: v.CashierName}
			items[i].AssignedAt = &assignedAt
		}
	}
	return items
}

func FromAssignmentView(v *service.AssignmentView) Assignment {
	a := v.Assignment
	result := Assignment{
		ID:     a.ID.String(),
		Status: a.Status,
		CashRegister: CashRegister{
			ID:     v.Register.ID.String(),
			Code:   v.Register.Code,
			Name:   v.Register.Name,
			Status: v.Register.Status,
		},
		Cashier:    Cashier{ID: a.UserID.String(), Name: v.CashierName},
		AssignedAt: formatTime(a.AssignedAt),
	}
	if a.ReleasedAt != nil {
		releasedAt := formatTime(*a.ReleasedAt)
		result.ReleasedAt = &releasedAt
	}
	if a.ReleasedBy != nil {
		releasedBy := a.ReleasedBy.String()
		result.ReleasedBy = &releasedBy
	}
	return result
}
