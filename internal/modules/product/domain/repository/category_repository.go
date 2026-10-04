package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Category, error)
	ExistsByName(ctx context.Context, companyID uuid.UUID, name string) (bool, error)
}
