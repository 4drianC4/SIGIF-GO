package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshTokenModel es el modelo de persistencia de refresh tokens.
type RefreshTokenModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	TokenHash string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	UserAgent string         `gorm:"type:varchar(500)"`
	IPAddress string         `gorm:"type:varchar(45)"`
	ExpiresAt time.Time      `gorm:"not null;index"`
	RevokedAt *time.Time     `gorm:"index"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}
