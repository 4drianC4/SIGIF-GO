package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryModel es el modelo de persistencia de categorías.
// El índice único es parcial para que una categoría borrada no bloquee su nombre.
type CategoryModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_categories_tenant_name,where:deleted_at IS NULL"`
	Name        string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_categories_tenant_name"`
	Description string         `gorm:"type:varchar(500)"`
	Status      string         `gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (CategoryModel) TableName() string {
	return "categories"
}
