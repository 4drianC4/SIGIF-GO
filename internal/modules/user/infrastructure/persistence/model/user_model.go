package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserModel es el modelo de persistencia de usuarios.
type UserModel struct {
	ID           uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID              `gorm:"type:uuid;not null;index:idx_users_tenant_email,unique"`
	Email        string                 `gorm:"type:varchar(255);not null;index:idx_users_tenant_email,unique"`
	PasswordHash string                 `gorm:"type:varchar(255);not null"`
	FirstName    string                 `gorm:"type:varchar(100);not null"`
	LastName     string                 `gorm:"type:varchar(100);not null"`
	Phone        string                 `gorm:"type:varchar(50)"`
	AvatarURL    string                 `gorm:"type:varchar(500)"`
	Roles        datatypes.JSONSlice[string] `gorm:"type:jsonb"`
	Status       string                 `gorm:"type:varchar(20);default:'pending';index"`
	LastLoginAt  *time.Time             `gorm:"index"`
	Settings     datatypes.JSON         `gorm:"type:jsonb"`
	CreatedAt    time.Time              `gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt         `gorm:"index"`
}

func (UserModel) TableName() string {
	return "users"
}