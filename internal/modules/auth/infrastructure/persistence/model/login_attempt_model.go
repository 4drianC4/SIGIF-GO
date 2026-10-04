package model

import (
	"time"

	"github.com/google/uuid"
)

type LoginAttemptModel struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email         string     `gorm:"type:varchar(160);not null;index"`
	CompanyID     *uuid.UUID `gorm:"type:uuid;index"`
	IPAddress     string     `gorm:"type:varchar(45);not null;index"`
	Successful    bool       `gorm:"not null"`
	FailureReason string     `gorm:"type:varchar(80)"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;index"`
}

func (LoginAttemptModel) TableName() string {
	return "login_attempt"
}
