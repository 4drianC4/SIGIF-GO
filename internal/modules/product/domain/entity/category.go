package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

type Category struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// NewCategory crea una categoría activa con el nombre normalizado.
func NewCategory(clock clock.Clock, tenantID uuid.UUID, name, description string) *Category {
	now := clock.NowUTC()
	return &Category{
		ID:          uuid.New(),
		TenantID:    tenantID,
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

// NormalizeName quita espacios sobrantes al inicio, al final y entre palabras.
func NormalizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}
