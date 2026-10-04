package model

import (
	"time"

	"github.com/google/uuid"
)

type RoleModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   *uuid.UUID `gorm:"type:uuid;index"`
	Name        string     `gorm:"type:varchar(60);not null;uniqueIndex:idx_roles_company_name"`
	Description string     `gorm:"type:varchar(200)"`
	IsTemplate  bool       `gorm:"not null;default:false"`
	IsSystem    bool       `gorm:"not null;default:false"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
}

func (RoleModel) TableName() string {
	return "role"
}
