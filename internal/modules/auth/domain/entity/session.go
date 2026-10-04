package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

type SessionCloseReason string

const (
	CloseReasonManual     SessionCloseReason = "manual"
	CloseReasonInactivity SessionCloseReason = "inactivity"
	CloseReasonRemote     SessionCloseReason = "remote_close"
	CloseReasonExpiration SessionCloseReason = "expiration"
)

// UserSession represents an authenticated session. Only the token hash is
// persisted so that reading the database never allows session hijacking.
type UserSession struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TokenHash   string
	IPAddress   string
	Device      string
	StartedAt   time.Time
	ExpiresAt   time.Time
	EndedAt     *time.Time
	CloseReason *SessionCloseReason
}

func NewSession(clock clock.Clock, id, userID uuid.UUID, tokenHash, ipAddress, device string, expiresIn time.Duration) *UserSession {
	now := clock.NowUTC()
	return &UserSession{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		IPAddress: ipAddress,
		Device:    device,
		StartedAt: now,
		ExpiresAt: now.Add(expiresIn),
	}
}

func (s *UserSession) IsActive(clock clock.Clock) bool {
	if s.EndedAt != nil {
		return false
	}
	return clock.NowUTC().Before(s.ExpiresAt)
}

func (s *UserSession) End(clock clock.Clock, reason SessionCloseReason) {
	now := clock.NowUTC()
	s.EndedAt = &now
	s.CloseReason = &reason
}
