package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
)

type CreateCategoryParams struct {
	CompanyID    uuid.UUID
	ParentID     *uuid.UUID
	Name         string
	Description  string
	DefaultTax   *string
	TargetMargin *decimal.Decimal
}

type CreatedCategory struct {
	Category       *entity.Category
	DefaultTaxName *string
}

func (s *CatalogService) CreateCategory(ctx context.Context, params CreateCategoryParams) (*CreatedCategory, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}

	if params.ParentID != nil {
		parent, err := s.categories.GetByID(ctx, params.CompanyID, *params.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrParentNotFound
		}
		if !parent.IsActive() {
			return nil, ErrCategoryInactive
		}
	}

	var taxID *uuid.UUID
	var taxName *string
	if params.DefaultTax != nil && strings.TrimSpace(*params.DefaultTax) != "" {
		tax, err := s.taxes.GetByName(ctx, strings.TrimSpace(*params.DefaultTax))
		if err != nil {
			return nil, err
		}
		if tax == nil {
			return nil, ErrTaxNotFound
		}
		taxID, taxName = &tax.ID, &tax.Name
	}

	category, err := entity.NewCategory(s.clock, entity.NewCategoryParams{
		CompanyID:    params.CompanyID,
		ParentID:     params.ParentID,
		Name:         params.Name,
		Description:  params.Description,
		DefaultTaxID: taxID,
		TargetMargin: params.TargetMargin,
	})
	if err != nil {
		return nil, err
	}

	exists, err := s.categories.ExistsByName(ctx, category.CompanyID, category.ParentID, category.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCategoryNameTaken
	}

	if err := s.categories.Create(ctx, category); err != nil {
		return nil, err
	}
	return &CreatedCategory{Category: category, DefaultTaxName: taxName}, nil
}

func (s *CatalogService) ListCategories(ctx context.Context, companyID uuid.UUID) ([]repository.CategoryListItem, error) {
	if companyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	return s.categories.List(ctx, companyID)
}
