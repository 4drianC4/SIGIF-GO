package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

// CategoryRepository es el puerto de persistencia de categorías.
type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.Category, error)
	// ExistsByName compara el nombre sin distinguir mayúsculas.
	ExistsByName(ctx context.Context, tenantID uuid.UUID, name string) (bool, error)
}
