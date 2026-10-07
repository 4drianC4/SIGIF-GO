package model

import (
	"time"

	"github.com/google/uuid"
)

type RoleModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   *uuid.UUID `gorm:"type:uuid;index;index:idx_roles_company_id_name"`
	Name        string     `gorm:"type:varchar(60);not null;uniqueIndex:idx_roles_company_name;index:idx_roles_company_id_name"`
	Description string     `gorm:"type:varchar(200)"`
	IsTemplate  bool       `gorm:"not null;default:false"`
	IsSystem    bool       `gorm:"not null;default:false"`
	Status      string     `gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
}

func (RoleModel) TableName() string {
	return "role"
}

// RoleListModel is the read model of the role listing: a role plus the number
// of permissions assigned to it, computed by the repository query. It is never
// migrated.
type RoleListModel struct {
	RoleModel
	PermissionsCount int64
}

func (RoleListModel) TableName() string {
	return "role"
}
