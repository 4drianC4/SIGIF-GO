package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type CreateCategoryParams struct {
	TenantID    uuid.UUID
	Name        string
	Description string
}

// CreateCategory registra una categoría validando que su nombre no exista en el tenant.
func (s *CatalogService) CreateCategory(ctx context.Context, params CreateCategoryParams) (*entity.Category, error) {
	if params.TenantID == uuid.Nil {
		return nil, sharedErrors.ErrTenantRequired
	}

	category := entity.NewCategory(s.clock, params.TenantID, params.Name, params.Description)

	exists, err := s.categories.ExistsByName(ctx, category.TenantID, category.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCategoryNameTaken
	}

	if err := s.categories.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}
