//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
	customerGorm "github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/customer/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func TestUpdateStatusWritesOnlyStatusAndUpdatedAt(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, _ := seedCustomers(t, db, repo)
	ctx := context.Background()
	found, _, _ := repo.List(ctx, companyA, repository.ListFilter{Q: "Supermercado"}, 0, 1)
	customer := found[0]

	// Another request edits the customer after it was read: the status change
	// must not write the stale values back.
	if err := db.DB.Model(&model.CustomerModel{}).Where("customer_id = ?", customer.ID).
		Updates(map[string]any{"legal_name": "Supermercado Norte SRL", "credit_balance": "75.50", "points_accrued": 12}).Error; err != nil {
		t.Fatalf("concurrent edit: %v", err)
	}
	updatedAt := time.Date(2026, 10, 13, 9, 30, 0, 0, time.UTC)

	if err := repo.UpdateStatus(ctx, companyA, customer.ID, entity.CustomerStatusInactive, updatedAt); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	stored, err := repo.GetByID(ctx, companyA, customer.ID)
	if err != nil || stored == nil {
		t.Fatalf("GetByID() = %v, %v", stored, err)
	}
	if stored.Status != entity.CustomerStatusInactive || stored.UpdatedAt == nil || !stored.UpdatedAt.Equal(updatedAt) {
		t.Errorf("status = %s, updated_at = %v, want inactive at %v", stored.Status, stored.UpdatedAt, updatedAt)
	}
	if stored.LegalName != "Supermercado Norte SRL" || stored.CreditBalance.String() != "75.5" || stored.PointsAccrued != 12 {
		t.Errorf("other columns were overwritten: %+v", stored)
	}
	if !stored.CreatedAt.Equal(customer.CreatedAt) || stored.DeletedAt != nil {
		t.Errorf("created_at/deleted_at changed: %+v", stored)
	}
}

func TestUpdateStatusNotFound(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, companyB := seedCustomers(t, db, repo)
	ctx := context.Background()
	found, _, _ := repo.List(ctx, companyA, repository.ListFilter{Q: "Supermercado"}, 0, 1)
	existing := found[0]

	var deleted model.CustomerModel
	if err := db.DB.Unscoped().Where("company_id = ? AND deleted_at IS NOT NULL", companyA).First(&deleted).Error; err != nil {
		t.Fatalf("deleted seed lookup: %v", err)
	}

	tests := []struct {
		name    string
		company uuid.UUID
		id      uuid.UUID
	}{
		{"missing", companyA, uuid.New()},
		{"foreign company", companyB, existing.ID},
		{"deleted", companyA, deleted.ID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateStatus(ctx, tt.company, tt.id, entity.CustomerStatusBlocked, time.Now().UTC())
			if !errors.Is(err, service.ErrCustomerNotFound) {
				t.Fatalf("UpdateStatus() error = %v, want ErrCustomerNotFound", err)
			}
		})
	}

	stored, _ := repo.GetByID(ctx, companyA, existing.ID)
	if stored.Status != entity.CustomerStatusActive {
		t.Errorf("status = %s, want active", stored.Status)
	}
	var after model.CustomerModel
	db.DB.Unscoped().Where("customer_id = ?", deleted.ID).First(&after)
	if after.Status != deleted.Status {
		t.Errorf("deleted row status = %s, want %s", after.Status, deleted.Status)
	}
}

func TestDeleteKeepsRowAndFreesDocument(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, companyB := seedCustomers(t, db, repo)
	ctx := context.Background()
	svc := service.NewCustomerService(repo, customerGorm.NewNoDependencyChecker(), clock.NewRealClock())
	found, _, _ := repo.List(ctx, companyA, repository.ListFilter{Q: "Supermercado"}, 0, 1)
	customer := found[0]
	before := countRows(t, db, companyA, companyB)

	if err := svc.Delete(ctx, companyA, customer.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if after := countRows(t, db, companyA, companyB); after != before {
		t.Errorf("rows = %d, want %d: the delete must be logical", after, before)
	}
	var row model.CustomerModel
	if err := db.DB.Unscoped().Where("customer_id = ?", customer.ID).First(&row).Error; err != nil {
		t.Fatalf("row lookup: %v", err)
	}
	if !row.DeletedAt.Valid || row.LegalName != "Supermercado Norte" {
		t.Errorf("row = %+v, want deleted_at set and data kept", row)
	}
	if stored, _ := repo.GetByID(ctx, companyA, customer.ID); stored != nil {
		t.Errorf("GetByID() = %+v, want nil after delete", stored)
	}
	if listed, total, _ := repo.List(ctx, companyA, repository.ListFilter{Q: "Supermercado"}, 0, 10); total != 0 || len(listed) != 0 {
		t.Errorf("deleted customer is still listed: %d", total)
	}
	if err := svc.Delete(ctx, companyA, customer.ID); !errors.Is(err, service.ErrCustomerNotFound) {
		t.Errorf("second Delete() error = %v, want ErrCustomerNotFound", err)
	}

	again := newCustomer(companyA, "Supermercado Nuevo", *customer.DocumentNumber)
	if err := repo.Create(ctx, again); err != nil {
		t.Errorf("document of a deleted customer must be reusable: %v", err)
	}
}

func TestDeleteRejectedChangesNothing(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, companyB := seedCustomers(t, db, repo)
	ctx := context.Background()
	deps := testutil.NewStubDependencyChecker()
	svc := service.NewCustomerService(repo, deps, clock.NewRealClock())
	found, _, _ := repo.List(ctx, companyA, repository.ListFilter{}, 0, 10)
	withRecords, withBalance, withPoints := found[0], found[1], found[2]

	deps.Block(withRecords.ID)
	db.DB.Model(&model.CustomerModel{}).Where("customer_id = ?", withBalance.ID).Update("credit_balance", "20.00")
	db.DB.Model(&model.CustomerModel{}).Where("customer_id = ?", withPoints.ID).Update("points_accrued", 5)

	tests := []struct {
		name string
		id   uuid.UUID
		want error
	}{
		{"related records", withRecords.ID, service.ErrCustomerHasDependencies},
		{"credit balance", withBalance.ID, service.ErrCustomerHasBalance},
		{"accrued points", withPoints.ID, service.ErrCustomerHasBalance},
		{"foreign company", withRecords.ID, service.ErrCustomerNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			company := companyA
			if tt.name == "foreign company" {
				company = companyB
			}
			var before model.CustomerModel
			db.DB.Unscoped().Where("customer_id = ?", tt.id).First(&before)

			if err := svc.Delete(ctx, company, tt.id); !errors.Is(err, tt.want) {
				t.Fatalf("Delete() error = %v, want %v", err, tt.want)
			}

			var after model.CustomerModel
			db.DB.Unscoped().Where("customer_id = ?", tt.id).First(&after)
			if after.DeletedAt.Valid || after.Status != before.Status || after.LegalName != before.LegalName ||
				(before.UpdatedAt == nil) != (after.UpdatedAt == nil) {
				t.Errorf("rejected delete changed the row:\n before %+v\n after  %+v", before, after)
			}
		})
	}
}
