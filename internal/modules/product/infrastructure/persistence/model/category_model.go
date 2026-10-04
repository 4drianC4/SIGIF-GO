package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_categories_company_name,where:deleted_at IS NULL"`
	Name        string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_categories_company_name"`
	Description string         `gorm:"type:varchar(500)"`
	Status      string         `gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (CategoryModel) TableName() string {
	return "categories"
}
