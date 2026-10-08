package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserModel is the persistence model for application users.
type UserModel struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID         *uuid.UUID     `gorm:"type:uuid;index"`
	RoleID            uuid.UUID      `gorm:"type:uuid;not null;index"`
	FirstName         string         `gorm:"type:varchar(80);not null"`
	LastName          string         `gorm:"type:varchar(80);not null"`
	Username          string         `gorm:"type:varchar(60);not null"`
	Email             string         `gorm:"type:varchar(160);not null;uniqueIndex"`
	Phone             string         `gorm:"type:varchar(30)"`
	PasswordHash      string         `gorm:"type:varchar(255);not null"`
	PasswordAlgorithm string         `gorm:"type:varchar(20);not null;default:'argon2id'"`
	RequiresOTP       bool           `gorm:"not null;default:false"`
	Status            string         `gorm:"type:varchar(20);not null;default:'active';index"`
	LastAccess        *time.Time     `gorm:"index"`
	CreatedAt         time.Time      `gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
	return "app_user"
}
