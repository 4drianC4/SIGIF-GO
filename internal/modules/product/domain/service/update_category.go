package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
)

// UpdateCategoryParams holds the fields of a partial update. DefaultTax is the
// name of a tax; clearing it, like the target margin, makes the category
// inherit the value of its parent.
type UpdateCategoryParams struct {
	CompanyID    uuid.UUID
	CategoryID   uuid.UUID
	Name         *string
	Description  *string
	DefaultTax   entity.Optional[string]
	TargetMargin entity.Optional[decimal.Decimal]
}

// UpdateCategory changes the fields that were sent and returns the category as
// the listing shows it. The name stays unique among the categories of the
// same level, ignoring case; the category itself does not count.
func (s *CatalogService) UpdateCategory(ctx context.Context, params UpdateCategoryParams) (*repository.CategoryListItem, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	category, err := s.categories.GetByID(ctx, params.CompanyID, params.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrCategoryNotFound
	}

	taxID := entity.Optional[uuid.UUID]{Set: params.DefaultTax.Set}
	if params.DefaultTax.Set && params.DefaultTax.Value != nil && strings.TrimSpace(*params.DefaultTax.Value) != "" {
		tax, err := s.taxes.GetByName(ctx, strings.TrimSpace(*params.DefaultTax.Value))
		if err != nil {
			return nil, err
		}
		if tax == nil {
			return nil, ErrTaxNotFound
		}
		taxID = entity.Some(tax.ID)
	}

	if err := category.Update(s.clock, entity.UpdateCategoryParams{
		Name:         params.Name,
		Description:  params.Description,
		DefaultTaxID: taxID,
		TargetMargin: params.TargetMargin,
	}); err != nil {
		return nil, err
	}

	if params.Name != nil {
		taken, err := s.categories.ExistsByName(ctx, category.CompanyID, category.ParentID, category.Name, category.ID)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, ErrCategoryNameTaken
		}
	}
	if err := s.categories.Update(ctx, category); err != nil {
		return nil, err
	}

	item, err := s.categories.GetListItem(ctx, category.CompanyID, category.ID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrCategoryNotFound
	}
	return item, nil
}
