package model

import (
	"time"

	"github.com/google/uuid"
)

type SessionModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash   string     `gorm:"type:varchar(64);not null;uniqueIndex"`
	IPAddress   string     `gorm:"type:varchar(45)"`
	Device      string     `gorm:"type:varchar(160)"`
	StartedAt   time.Time  `gorm:"not null"`
	ExpiresAt   time.Time  `gorm:"not null;index"`
	EndedAt     *time.Time `gorm:"index"`
	CloseReason *string    `gorm:"type:varchar(20)"`
}

func (SessionModel) TableName() string {
	return "user_session"
}
