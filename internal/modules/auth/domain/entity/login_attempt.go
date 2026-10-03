package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

// LoginAttempt is the audit trail for authentication attempts (successful or
// not), used for traceability and future lockout support.
type LoginAttempt struct {
	ID            uuid.UUID
	Email         string
	CompanyID     *uuid.UUID
	IPAddress     string
	Successful    bool
	FailureReason string
	CreatedAt     time.Time
}

func NewLoginAttempt(clock clock.Clock, email string, companyID *uuid.UUID, ipAddress string, successful bool, failureReason string) *LoginAttempt {
	return &LoginAttempt{
		ID:            uuid.New(),
		Email:         email,
		CompanyID:     companyID,
		IPAddress:     ipAddress,
		Successful:    successful,
		FailureReason: failureReason,
		CreatedAt:     clock.NowUTC(),
	}
}
