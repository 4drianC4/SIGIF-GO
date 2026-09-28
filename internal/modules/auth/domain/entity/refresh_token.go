package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type RefreshToken struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	TokenHash string     `json:"-" gorm:"type:varchar(255);not null;uniqueIndex"`
	UserAgent string     `json:"user_agent" gorm:"type:varchar(500)"`
	IPAddress string     `json:"ip_address" gorm:"type:varchar(45)"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null;index"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func NewRefreshToken(clock clock.Clock, userID uuid.UUID, tokenHash, userAgent, ipAddress string, expiresIn time.Duration) *RefreshToken {
	now := clock.NowUTC()
	return &RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: now.Add(expiresIn),
		CreatedAt: now,
	}
}

func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

func (r *RefreshToken) IsRevoked() bool {
	return r.RevokedAt != nil
}

func (r *RefreshToken) Revoke(clock clock.Clock) {
	now := clock.NowUTC()
	r.RevokedAt = &now
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}