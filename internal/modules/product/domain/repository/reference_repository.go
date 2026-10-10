package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type UnitOfMeasureRepository interface {
	List(ctx context.Context) ([]*entity.UnitOfMeasure, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type TaxRepository interface {
	List(ctx context.Context) ([]*entity.Tax, error)
	GetByName(ctx context.Context, name string) (*entity.Tax, error)
}
