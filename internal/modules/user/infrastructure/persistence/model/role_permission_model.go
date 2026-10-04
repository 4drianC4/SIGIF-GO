package model

import "github.com/google/uuid"

type RolePermissionModel struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (RolePermissionModel) TableName() string {
	return "role_permission"
}
