package command

import (
	"github.com/google/uuid"
)

type AssignCashRegister struct {
	CompanyID      uuid.UUID
	UserID         uuid.UUID
	CashRegisterID uuid.UUID
}

type ReleaseAssignment struct {
	CompanyID     uuid.UUID
	ActorID       uuid.UUID
	AssignmentID  uuid.UUID
	CanReleaseAny bool
}
