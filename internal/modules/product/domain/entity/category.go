package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

type Category struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Name        string
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func NewCategory(clock clock.Clock, companyID uuid.UUID, name, description string) *Category {
	now := clock.NowUTC()
	return &Category{
		ID:          uuid.New(),
		CompanyID:   companyID,
		Name:        NormalizeName(name),
		Description: strings.TrimSpace(description),
		Status:      StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (c *Category) IsActive() bool {
	return c.Status == StatusActive
}

func NormalizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}
