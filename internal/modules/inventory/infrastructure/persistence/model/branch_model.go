package model

import (
	"time"

	"github.com/google/uuid"
)

type BranchModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null"`
	Name      string    `gorm:"type:varchar(150);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (BranchModel) TableName() string { return "branches" }

type LocationModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	BranchID  int64     `gorm:"not null"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null"`
	Name      string    `gorm:"type:varchar(150);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (LocationModel) TableName() string { return "locations" }
