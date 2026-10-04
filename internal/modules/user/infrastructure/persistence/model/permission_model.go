package model

import "github.com/google/uuid"

type PermissionModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Module      string    `gorm:"type:varchar(60);not null"`
	Operation   string    `gorm:"type:varchar(60);not null"`
	Description string    `gorm:"type:varchar(200)"`
}

func (PermissionModel) TableName() string {
	return "permission"
}
