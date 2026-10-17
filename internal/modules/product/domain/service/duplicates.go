package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// findConflicts returns every value that another product of the company,
// other than excludeID, already holds (HU-06-01). The values arrive
// normalized; empty values are never duplicated.
func (s *CatalogService) findConflicts(ctx context.Context, companyID uuid.UUID, name, sku, barcode string, excludeID uuid.UUID) ([]DuplicateConflict, error) {
	checks := []struct {
		field string
		value string
		find  func(context.Context, uuid.UUID, string, uuid.UUID) (*entity.Product, error)
	}{
		{"name", name, s.products.FindByName},
		{"sku", sku, s.products.FindBySKU},
		{"barcode", barcode, s.products.FindByBarcode},
	}
	var conflicts []DuplicateConflict
	for _, check := range checks {
		if check.value == "" {
			continue
		}
		existing, err := check.find(ctx, companyID, check.value, excludeID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			conflicts = append(conflicts, newDuplicateConflict(check.field, check.value, existing))
		}
	}
	return conflicts, nil
}

// detailDuplicate turns a unique index violation into the detailed conflict.
// It happens when a concurrent request takes the value after the checks
// passed; the winner is looked up again outside the failed write.
func (s *CatalogService) detailDuplicate(ctx context.Context, err error, companyID uuid.UUID, name, sku, barcode string, excludeID uuid.UUID) error {
	var appErr *sharedErrors.AppError
	if errors.As(err, &appErr) {
		if _, detailed := appErr.Details.(DuplicateDetails); detailed {
			return err
		}
	}
	if !errors.Is(err, ErrProductNameTaken) && !errors.Is(err, ErrSKUTaken) && !errors.Is(err, ErrBarcodeTaken) {
		return err
	}
	conflicts, findErr := s.findConflicts(ctx, companyID, name, sku, barcode, excludeID)
	if findErr != nil || len(conflicts) == 0 {
		return err
	}
	return newDuplicateError(conflicts)
}
