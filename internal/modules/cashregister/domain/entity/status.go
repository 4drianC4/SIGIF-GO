package entity

type CashRegisterStatus string

const (
	CashRegisterStatusActive   CashRegisterStatus = "active"
	CashRegisterStatusInactive CashRegisterStatus = "inactive"
)

type AssignmentStatus string

const (
	AssignmentStatusActive   AssignmentStatus = "active"
	AssignmentStatusReleased AssignmentStatus = "released"
)

// Availability tells whether a cash register is free or held by a cashier.
type Availability string

const (
	AvailabilityAvailable Availability = "available"
	AvailabilityInUse     Availability = "in_use"
)

func (a Availability) IsValid() bool {
	return a == AvailabilityAvailable || a == AvailabilityInUse
}
