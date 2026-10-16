package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
)

func (s *SupplierService) Get(ctx context.Context, companyID, id uuid.UUID) (*entity.Supplier, error) {
	if companyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	supplier, err := s.repo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, ErrSupplierNotFound
	}
	return supplier, nil
}

func (s *SupplierService) Create(ctx context.Context, companyID uuid.UUID, fields entity.Fields) (*entity.Supplier, error) {
	if companyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	supplier, err := entity.NewSupplier(s.clock, companyID, fields)
	if err != nil {
		return nil, err
	}
	if err := s.checkTaxID(ctx, supplier, uuid.Nil); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, supplier); err != nil {
		return nil, err
	}
	return supplier, nil
}

func (s *SupplierService) Update(ctx context.Context, companyID, id uuid.UUID, fields entity.Fields) (*entity.Supplier, error) {
	fields.Normalize()
	if err := fields.Validate(); err != nil {
		return nil, err
	}
	var updated *entity.Supplier
	err := s.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		current, err := s.Get(ctx, companyID, id)
		if err != nil {
			return err
		}
		current.Fields = fields
		if err := s.checkTaxID(ctx, current, id); err != nil {
			return err
		}
		now := s.clock.NowUTC()
		current.UpdatedAt = &now
		if err := s.repo.Update(ctx, current); err != nil {
			return err
		}
		updated = current
		return nil
	})
	return updated, err
}

func (s *SupplierService) checkTaxID(ctx context.Context, supplier *entity.Supplier, excludeID uuid.UUID) error {
	if supplier.TaxID == nil {
		return nil
	}
	exists, err := s.repo.ExistsByTaxID(ctx, supplier.CompanyID, *supplier.TaxID, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return ErrTaxIDTaken
	}
	return nil
}

// Deactivation is idempotent and preserves the supplier and its associations.
func (s *SupplierService) Deactivate(ctx context.Context, companyID, id uuid.UUID) (*entity.Supplier, error) {
	var result *entity.Supplier
	err := s.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		current, err := s.Get(ctx, companyID, id)
		if err != nil {
			return err
		}
		if current.Status != entity.StatusInactive {
			now := s.clock.NowUTC()
			if err := s.repo.Deactivate(ctx, companyID, id, now); err != nil {
				return err
			}
			current.Status, current.UpdatedAt = entity.StatusInactive, &now
		}
		result = current
		return nil
	})
	return result, err
}

func (s *SupplierService) List(ctx context.Context, companyID uuid.UUID, filter repository.ListFilter, offset, limit int) ([]repository.SupplierListItem, int64, error) {
	if companyID == uuid.Nil {
		return nil, 0, ErrCompanyRequired
	}
	if filter.Status != nil && *filter.Status != entity.StatusActive && *filter.Status != entity.StatusInactive {
		return nil, 0, entity.Invalid("status", "must be active or inactive")
	}
	return s.repo.List(ctx, companyID, filter, offset, limit)
}

func (s *SupplierService) FormOptions(ctx context.Context, companyID uuid.UUID) (repository.FormOptions, error) {
	if companyID == uuid.Nil {
		return repository.FormOptions{}, ErrCompanyRequired
	}
	return s.repo.FormOptions(ctx, companyID)
}
