package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/service"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func ptr[T any](v T) *T { return &v }

type key struct{ supplier, product uuid.UUID }
type storedProduct struct {
	company uuid.UUID
	ref     repository.ProductRef
}
type memoryAssociations struct {
	repository.ProductSupplierRepository
	products    map[uuid.UUID]storedProduct
	rows        map[key]*entity.ProductSupplier
	createError error
}
type memorySuppliers struct {
	repository.SupplierRepository
	rows         map[uuid.UUID]*entity.Supplier
	associations *memoryAssociations
}

func (m *memorySuppliers) GetByID(_ context.Context, companyID, id uuid.UUID) (*entity.Supplier, error) {
	s := m.rows[id]
	if s == nil || s.CompanyID != companyID {
		return nil, nil
	}
	copy := *s
	return &copy, nil
}
func (m *memorySuppliers) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	before := make(map[key]*entity.ProductSupplier, len(m.associations.rows))
	for k, a := range m.associations.rows {
		copy := *a
		before[k] = &copy
	}
	if err := fn(ctx); err != nil {
		m.associations.rows = before
		return err
	}
	return nil
}
func (m *memoryAssociations) GetProduct(_ context.Context, companyID, id uuid.UUID) (*repository.ProductRef, error) {
	p, ok := m.products[id]
	if !ok || p.company != companyID {
		return nil, nil
	}
	copy := p.ref
	return &copy, nil
}
func (m *memoryAssociations) Get(_ context.Context, companyID, supplierID, productID uuid.UUID) (*entity.ProductSupplier, error) {
	a := m.rows[key{supplierID, productID}]
	if a == nil || a.CompanyID != companyID {
		return nil, nil
	}
	copy := *a
	return &copy, nil
}
func (m *memoryAssociations) CreateMany(_ context.Context, items []*entity.ProductSupplier) error {
	if m.createError != nil {
		return m.createError
	}
	for _, a := range items {
		if m.rows[key{a.SupplierID, a.ProductID}] != nil {
			return service.ErrAlreadyAssociated
		}
	}
	for _, a := range items {
		copy := *a
		m.rows[key{a.SupplierID, a.ProductID}] = &copy
	}
	return nil
}
func (m *memoryAssociations) Update(_ context.Context, a *entity.ProductSupplier) error {
	k := key{a.SupplierID, a.ProductID}
	if m.rows[k] == nil || m.rows[k].CompanyID != a.CompanyID {
		return service.ErrAssociationNotFound
	}
	copy := *a
	m.rows[k] = &copy
	return nil
}
func (m *memoryAssociations) ClearPreferred(_ context.Context, companyID, productID uuid.UUID, at time.Time) error {
	for _, a := range m.rows {
		if a.CompanyID == companyID && a.ProductID == productID && a.Preferred {
			a.Preferred = false
			a.UpdatedAt = &at
		}
	}
	return nil
}
func (m *memoryAssociations) Delete(_ context.Context, companyID, supplierID, productID uuid.UUID) error {
	k := key{supplierID, productID}
	if m.rows[k] == nil || m.rows[k].CompanyID != companyID {
		return service.ErrAssociationNotFound
	}
	delete(m.rows, k)
	return nil
}

type fixture struct {
	svc                                       *service.SupplierService
	company, supplier, product, secondProduct uuid.UUID
	suppliers                                 *memorySuppliers
	associations                              *memoryAssociations
}

func newFixture() fixture {
	f := fixture{company: uuid.New(), supplier: uuid.New(), product: uuid.New(), secondProduct: uuid.New()}
	f.associations = &memoryAssociations{products: map[uuid.UUID]storedProduct{}, rows: map[key]*entity.ProductSupplier{}}
	for _, id := range []uuid.UUID{f.product, f.secondProduct} {
		f.associations.products[id] = storedProduct{company: f.company, ref: repository.ProductRef{ID: id, Name: "Producto", Status: "active"}}
	}
	f.suppliers = &memorySuppliers{rows: map[uuid.UUID]*entity.Supplier{f.supplier: {ID: f.supplier, CompanyID: f.company, Status: entity.StatusActive}}, associations: f.associations}
	f.svc = service.NewSupplierService(f.suppliers, f.associations, clock.NewMockClock(time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)))
	return f
}
func (f fixture) params(supplier, product uuid.UUID) service.AssociationParams {
	return service.AssociationParams{CompanyID: f.company, SupplierID: supplier, ProductID: product, CurrentCost: ptr(decimal.RequireFromString("13.10")), LeadTimeDays: ptr(3)}
}

func TestBatchAssociationsAreAtomicAndInitiallyIncomplete(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	items, err := f.svc.AssociateProducts(ctx, f.company, f.supplier, []uuid.UUID{f.secondProduct, f.product})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].ProductID != f.secondProduct {
		t.Fatal("selection order changed")
	}
	for _, a := range items {
		if a.CurrentCost != nil || a.SupplierCode != nil || a.Preferred || a.LeadTimeDays != nil {
			t.Fatalf("bulk association must leave optional values unset: %+v", a)
		}
	}
	third := uuid.New()
	f.associations.products[third] = storedProduct{company: f.company, ref: repository.ProductRef{ID: third, Status: "active"}}
	_, err = f.svc.AssociateProducts(ctx, f.company, f.supplier, []uuid.UUID{third, f.product})
	if !errors.Is(err, service.ErrAlreadyAssociated) || len(f.associations.rows) != 2 {
		t.Fatalf("duplicate batch was not rejected atomically: %v", err)
	}
	_, err = f.svc.AssociateProducts(ctx, f.company, f.supplier, []uuid.UUID{third, uuid.New()})
	if !errors.Is(err, service.ErrProductNotFound) || len(f.associations.rows) != 2 {
		t.Fatal("missing product caused partial creation")
	}
	_, err = f.svc.AssociateProducts(ctx, f.company, f.supplier, []uuid.UUID{third, third})
	if !sharedErrors.Is(err, sharedErrors.CodeValidation) {
		t.Fatalf("duplicate IDs accepted: %v", err)
	}
	if err := f.svc.Unlink(ctx, f.company, f.supplier, f.product); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.associations.products[f.product]; !ok {
		t.Fatal("unlink deleted product")
	}
	if f.suppliers.rows[f.supplier] == nil {
		t.Fatal("unlink deleted supplier")
	}
	if err := f.svc.Unlink(ctx, f.company, f.supplier, f.product); !errors.Is(err, service.ErrAssociationNotFound) {
		t.Fatalf("second unlink: %v", err)
	}
}

func TestPreferredSupplierSwitchAndRollback(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	first := f.params(f.supplier, f.product)
	first.Preferred = ptr(true)
	if _, err := f.svc.Associate(ctx, first); err != nil {
		t.Fatal(err)
	}
	other := uuid.New()
	f.suppliers.rows[other] = &entity.Supplier{ID: other, CompanyID: f.company, Status: entity.StatusActive}
	second := f.params(other, f.product)
	second.Preferred = ptr(true)
	// Simulate an insert failure after the former preference has been cleared.
	f.associations.createError = errors.New("insert failed")
	if _, err := f.svc.Associate(ctx, second); err == nil {
		t.Fatal("expected persistence error")
	}
	if !f.associations.rows[key{f.supplier, f.product}].Preferred || len(f.associations.rows) != 1 {
		t.Fatal("failed insert lost original preference")
	}
	f.associations.createError = nil
	a, err := f.svc.Associate(ctx, second)
	if err != nil {
		t.Fatal(err)
	}
	if !a.Preferred || f.associations.rows[key{f.supplier, f.product}].Preferred {
		t.Fatal("preferred supplier was not switched atomically")
	}
	second.CurrentCost, second.LeadTimeDays, second.SupplierCode, second.Preferred = ptr(decimal.Zero), ptr(0), ptr(" CODE-001 "), nil
	updated, err := f.svc.UpdateAssociation(ctx, second)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Preferred || updated.CurrentCost == nil || !updated.CurrentCost.IsZero() || *updated.LeadTimeDays != 0 || *updated.SupplierCode != "CODE-001" || updated.ID != a.ID || !updated.CreatedAt.Equal(a.CreatedAt) {
		t.Fatalf("update: %+v", updated)
	}
	second.CurrentCost, second.SupplierCode = nil, ptr(" ")
	updated, err = f.svc.UpdateAssociation(ctx, second)
	if err != nil || updated.CurrentCost != nil || updated.SupplierCode != nil {
		t.Fatalf("optional values not cleared: %+v %v", updated, err)
	}
	if err := f.svc.Unlink(ctx, f.company, other, f.product); err != nil {
		t.Fatal(err)
	}
	if f.associations.rows[key{f.supplier, f.product}].Preferred {
		t.Fatal("unlink silently chose another preferred supplier")
	}
}

func TestAssociationOwnershipStatusAndValidation(t *testing.T) {
	ctx := context.Background()
	for _, name := range []string{"foreign company", "foreign supplier", "foreign product", "inactive supplier", "inactive product", "negative cost", "too precise", "long code", "nil cost"} {
		t.Run(name, func(t *testing.T) {
			f := newFixture()
			p := f.params(f.supplier, f.product)
			want := sharedErrors.CodeValidation
			switch name {
			case "foreign company":
				p.CompanyID = uuid.New()
				want = sharedErrors.CodeNotFound
			case "foreign supplier":
				p.SupplierID = uuid.New()
				want = sharedErrors.CodeNotFound
			case "foreign product":
				f.associations.products[f.product] = storedProduct{company: uuid.New(), ref: repository.ProductRef{ID: f.product, Status: "active"}}
				want = sharedErrors.CodeNotFound
			case "inactive supplier":
				f.suppliers.rows[f.supplier].Status = entity.StatusInactive
				want = sharedErrors.CodeConflict
			case "inactive product":
				v := f.associations.products[f.product]
				v.ref.Status = "inactive"
				f.associations.products[f.product] = v
				want = sharedErrors.CodeConflict
			case "negative cost":
				p.CurrentCost = ptr(decimal.RequireFromString("-0.01"))
			case "too precise":
				p.CurrentCost = ptr(decimal.RequireFromString("1.001"))
			case "long code":
				p.SupplierCode = ptr("1234567890123456789012345678901234567890123456789012345678901")
			case "nil cost":
				p.CurrentCost = nil
			}
			_, err := f.svc.Associate(ctx, p)
			if !sharedErrors.Is(err, want) || len(f.associations.rows) != 0 {
				t.Fatalf("wanted %s with no writes, got %v", want, err)
			}
		})
	}
	f := newFixture()
	if _, err := f.svc.UpdateAssociation(ctx, f.params(f.supplier, f.product)); !errors.Is(err, service.ErrAssociationNotFound) {
		t.Fatalf("update inserted a missing association: %v", err)
	}
}
