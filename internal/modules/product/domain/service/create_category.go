package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type CreateCategoryParams struct {
	CompanyID   uuid.UUID
	Name        string
	Description string
}

func (s *CatalogService) CreateCategory(ctx context.Context, params CreateCategoryParams) (*entity.Category, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}

	category := entity.NewCategory(s.clock, params.CompanyID, params.Name, params.Description)

	exists, err := s.categories.ExistsByName(ctx, category.CompanyID, category.Name)
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
