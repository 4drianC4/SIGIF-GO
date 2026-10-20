package service

import (
	"context"

	"github.com/google/uuid"
)

// DeleteCategory marks a category as deleted. It is refused while the category
// still has subcategories or products, whatever their status: with a logical
// delete the foreign keys do not stop it, so the service does.
//
// The category row is locked for the whole transaction. A product or
// subcategory being created in it holds a shared lock on that row, so either
// it is counted here or it finds the category already deleted.
func (s *CatalogService) DeleteCategory(ctx context.Context, companyID, categoryID uuid.UUID) error {
	if companyID == uuid.Nil {
		return ErrCompanyRequired
	}
	return s.categories.WithinTransaction(ctx, func(txCtx context.Context) error {
		category, err := s.categories.GetByIDForUpdate(txCtx, companyID, categoryID)
		if err != nil {
			return err
		}
		if category == nil {
			return ErrCategoryNotFound
		}

		subcategories, err := s.categories.CountSubcategories(txCtx, companyID, categoryID)
		if err != nil {
			return err
		}
		if subcategories > 0 {
			return ErrCategoryHasSubcategories
		}
		products, err := s.categories.CountProducts(txCtx, companyID, categoryID)
		if err != nil {
			return err
		}
		if products > 0 {
			return ErrCategoryHasProducts
		}

		deleted, err := s.categories.SoftDelete(txCtx, companyID, categoryID)
		if err != nil {
			return err
		}
		if !deleted {
			return ErrCategoryNotFound
		}
		return nil
	})
}
