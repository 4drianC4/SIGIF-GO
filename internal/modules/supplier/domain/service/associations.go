package service

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
)

type AssociationParams struct {
	CompanyID, SupplierID, ProductID uuid.UUID
	SupplierCode                     *string
	LeadTimeDays                     *int
	CurrentCost                      *decimal.Decimal
	Preferred                        *bool
}

func (s *SupplierService) product(ctx context.Context, companyID, id uuid.UUID) (*repository.ProductRef, error) {
	if companyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	p, err := s.products.GetProduct(ctx, companyID, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	return p, nil
}

func (s *SupplierService) Associate(ctx context.Context, p AssociationParams) (*entity.ProductSupplier, error) {
	if p.CurrentCost == nil {
		return nil, entity.Invalid("current_cost", "required")
	}
	now := s.clock.NowUTC()
	a := &entity.ProductSupplier{ID: uuid.New(), CompanyID: p.CompanyID, SupplierID: p.SupplierID,
		ProductID: p.ProductID, SupplierCode: p.SupplierCode, LeadTimeDays: p.LeadTimeDays, CurrentCost: p.CurrentCost, CreatedAt: now}
	if p.Preferred != nil {
		a.Preferred = *p.Preferred
	}
	if err := a.NormalizeAndValidate(); err != nil {
		return nil, err
	}
	err := s.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		// Every writer locks product(s) first, then supplier, in the same order.
		product, err := s.product(ctx, p.CompanyID, p.ProductID)
		if err != nil {
			return err
		}
		supplier, err := s.Get(ctx, p.CompanyID, p.SupplierID)
		if err != nil {
			return err
		}
		if supplier.Status != entity.StatusActive {
			return ErrInactiveSupplier
		}
		if product.Status != "active" {
			return ErrInactiveProduct
		}
		existing, err := s.products.Get(ctx, p.CompanyID, p.SupplierID, p.ProductID)
		if err != nil {
			return err
		}
		if existing != nil {
			return ErrAlreadyAssociated
		}
		if a.Preferred {
			if err := s.products.ClearPreferred(ctx, p.CompanyID, p.ProductID, now); err != nil {
				return err
			}
		}
		return s.products.CreateMany(ctx, []*entity.ProductSupplier{a})
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// AssociateProducts creates all selected associations atomically. Missing cost
// and supplier code remain NULL until the user edits the relationship.
func (s *SupplierService) AssociateProducts(ctx context.Context, companyID, supplierID uuid.UUID, ids []uuid.UUID) ([]*entity.ProductSupplier, error) {
	if companyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	if len(ids) == 0 || len(ids) > 100 {
		return nil, entity.Invalid("product_ids", "select between 1 and 100 products")
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if id == uuid.Nil || seen[id] {
			return nil, entity.Invalid("product_ids", "must contain distinct, non-empty UUIDs")
		}
		seen[id] = true
	}
	// Stable lock order prevents deadlocks between overlapping selections.
	ordered := append([]uuid.UUID(nil), ids...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].String() < ordered[j].String() })
	items := make([]*entity.ProductSupplier, 0, len(ids))
	err := s.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		for _, id := range ordered {
			product, err := s.product(ctx, companyID, id)
			if err != nil {
				return err
			}
			if product.Status != "active" {
				return ErrInactiveProduct
			}
		}
		supplier, err := s.Get(ctx, companyID, supplierID)
		if err != nil {
			return err
		}
		if supplier.Status != entity.StatusActive {
			return ErrInactiveSupplier
		}
		for _, id := range ids {
			existing, err := s.products.Get(ctx, companyID, supplierID, id)
			if err != nil {
				return err
			}
			if existing != nil {
				return ErrAlreadyAssociated
			}
			items = append(items, &entity.ProductSupplier{ID: uuid.New(), CompanyID: companyID, SupplierID: supplierID, ProductID: id, CreatedAt: s.clock.NowUTC()})
		}
		return s.products.CreateMany(ctx, items)
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *SupplierService) UpdateAssociation(ctx context.Context, p AssociationParams) (*entity.ProductSupplier, error) {
	var result *entity.ProductSupplier
	err := s.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.product(ctx, p.CompanyID, p.ProductID); err != nil {
			return err
		}
		supplier, err := s.Get(ctx, p.CompanyID, p.SupplierID)
		if err != nil {
			return err
		}
		a, err := s.products.Get(ctx, p.CompanyID, p.SupplierID, p.ProductID)
		if err != nil {
			return err
		}
		if a == nil {
			return ErrAssociationNotFound
		}
		a.SupplierCode, a.LeadTimeDays, a.CurrentCost = p.SupplierCode, p.LeadTimeDays, p.CurrentCost
		if p.Preferred != nil {
			if *p.Preferred && supplier.Status != entity.StatusActive {
				return ErrInactiveSupplier
			}
			a.Preferred = *p.Preferred
		}
		if err := a.NormalizeAndValidate(); err != nil {
			return err
		}
		now := s.clock.NowUTC()
		a.UpdatedAt = &now
		if a.Preferred {
			if err := s.products.ClearPreferred(ctx, p.CompanyID, p.ProductID, now); err != nil {
				return err
			}
		}
		if err := s.products.Update(ctx, a); err != nil {
			return err
		}
		result = a
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SupplierService) Unlink(ctx context.Context, companyID, supplierID, productID uuid.UUID) error {
	return s.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.product(ctx, companyID, productID); err != nil {
			return err
		}
		if _, err := s.Get(ctx, companyID, supplierID); err != nil {
			return err
		}
		// Deleting a preferred association does not silently choose a replacement.
		return s.products.Delete(ctx, companyID, supplierID, productID)
	})
}

func (s *SupplierService) ListProductSuppliers(ctx context.Context, companyID, productID uuid.UUID, search string, offset, limit int) ([]repository.SupplierItem, int64, error) {
	if _, err := s.product(ctx, companyID, productID); err != nil {
		return nil, 0, err
	}
	return s.products.ListSuppliers(ctx, companyID, productID, search, offset, limit)
}

func (s *SupplierService) ProductOptions(ctx context.Context, companyID, supplierID uuid.UUID, filter repository.ProductOptionFilter, offset, limit int) ([]repository.ProductOption, int64, error) {
	supplier, err := s.Get(ctx, companyID, supplierID)
	if err != nil {
		return nil, 0, err
	}
	items, total, err := s.products.ProductOptions(ctx, companyID, supplierID, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		items[i].Selectable = !items[i].IsAssociated && items[i].Status == "active" && supplier.Status == entity.StatusActive
	}
	return items, total, nil
}

func (s *SupplierService) SupplierOptions(ctx context.Context, companyID, productID uuid.UUID, search string, offset, limit int) ([]repository.SupplierOption, int64, error) {
	product, err := s.product(ctx, companyID, productID)
	if err != nil {
		return nil, 0, err
	}
	items, total, err := s.products.SupplierOptions(ctx, companyID, productID, search, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		items[i].Selectable = !items[i].IsAssociated && items[i].Supplier.Status == entity.StatusActive && product.Status == "active"
	}
	return items, total, nil
}
