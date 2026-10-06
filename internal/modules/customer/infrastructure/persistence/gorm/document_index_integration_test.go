//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
	customerGorm "github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func TestDocumentUniqueIndex(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, companyB := seedCustomers(t, db, repo)
	ctx := context.Background()

	t.Run("duplicate document in the same company is a conflict", func(t *testing.T) {
		err := repo.Create(ctx, newCustomer(companyA, "Duplicado", "1020304050"))
		if !errors.Is(err, service.ErrDocumentTaken) {
			t.Fatalf("Create() error = %v, want ErrDocumentTaken", err)
		}
	})

	t.Run("same number with another document type is allowed", func(t *testing.T) {
		c := newCustomer(companyA, "Pasaporte", "1020304050")
		c.DocumentType = entity.DocumentTypePassport
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	})

	t.Run("same document in another company is allowed", func(t *testing.T) {
		if err := repo.Create(ctx, newCustomer(companyB, "Otra", "7788990011")); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	})

	t.Run("customers without document do not collide", func(t *testing.T) {
		for _, name := range []string{"Sin documento 1", "Sin documento 2"} {
			if err := repo.Create(ctx, newCustomer(companyA, name, "")); err != nil {
				t.Fatalf("Create(%s) error = %v", name, err)
			}
		}
	})

	t.Run("document of a deleted customer can be reused", func(t *testing.T) {
		if err := repo.Create(ctx, newCustomer(companyA, "Reusa eliminado", "9999999999")); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	})

	t.Run("update that bypasses the service check is a conflict", func(t *testing.T) {
		found, _, _ := repo.List(ctx, companyA, repository.ListFilter{Q: "Supermercado"}, 0, 1)
		customer := found[0]
		customer.Update(clock.NewRealClock(), customer.LegalName, entity.DocumentTypeTaxID, strPtr("1020304050"), nil, nil, nil)

		if err := repo.Update(ctx, customer); !errors.Is(err, service.ErrDocumentTaken) {
			t.Fatalf("Update() error = %v, want ErrDocumentTaken", err)
		}
		stored, _ := repo.GetByID(ctx, companyA, customer.ID)
		if stored.DocumentNumber == nil || *stored.DocumentNumber != "7788990011" {
			t.Errorf("document_number = %v, the row must be unchanged", stored.DocumentNumber)
		}
	})

}
