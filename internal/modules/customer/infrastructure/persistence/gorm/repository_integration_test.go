//go:build integration

package gorm_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	gormlib "gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	customerGorm "github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

func openDatabase(t *testing.T) *sharedDatabase.Database {
	t.Helper()
	dsn := os.Getenv("SIGIF_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SIGIF_TEST_DATABASE_DSN no está definido")
	}
	db, err := gormlib.Open(postgres.Open(dsn), &gormlib.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.CustomerModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return &sharedDatabase.Database{DB: db}
}

func strPtr(s string) *string { return &s }

func seedCustomers(t *testing.T, db *sharedDatabase.Database, repo repository.CustomerRepository) (uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	companyA, companyB := uuid.New(), uuid.New()
	t.Cleanup(func() {
		db.DB.Unscoped().Where("company_id IN ?", []uuid.UUID{companyA, companyB}).Delete(&model.CustomerModel{})
	})

	rows := []struct {
		company uuid.UUID
		name    string
		doc     string
		email   string
		status  entity.CustomerStatus
		deleted bool
	}{
		{companyA, "Ferretería Central", "1020304050", "central@ferre.com", entity.CustomerStatusActive, false},
		{companyA, "Supermercado Norte", "7788990011", "norte@super.com", entity.CustomerStatusActive, false},
		{companyA, "Farmacia 100% Sur", "5566778899", "sur@farma.com", entity.CustomerStatusActive, false},
		{companyA, "Ferretería Cerrada", "1020300000", "cerrada@ferre.com", entity.CustomerStatusInactive, false},
		{companyA, "Ferretería Eliminada", "9999999999", "eliminada@ferre.com", entity.CustomerStatusActive, true},
		{companyB, "Ferretería Central", "1020304050", "central@ferre.com", entity.CustomerStatusActive, false},
	}
	for i, row := range rows {
		clk := clock.NewMockClock(time.Date(2026, 10, 1, 8, i, 0, 0, time.UTC))
		c := entity.NewCustomer(clk, row.company, row.name, entity.DocumentTypeTaxID, strPtr(row.doc), nil, strPtr(row.email))
		c.Status = row.status
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("create %s: %v", row.name, err)
		}
		if row.deleted {
			c.SoftDelete(clk)
			if err := repo.Update(ctx, c); err != nil {
				t.Fatalf("soft delete %s: %v", row.name, err)
			}
		}
	}
	return companyA, companyB
}

func TestListSearchesWithinCompany(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, companyB := seedCustomers(t, db, repo)
	ctx := context.Background()
	active := entity.CustomerStatusActive
	inactive := entity.CustomerStatusInactive

	tests := []struct {
		name    string
		company uuid.UUID
		filter  repository.ListFilter
		want    []string
	}{
		{name: "active only", company: companyA, filter: repository.ListFilter{Status: &active, SortBy: repository.SortByLegalName, SortOrder: repository.SortAsc},
			want: []string{"Farmacia 100% Sur", "Ferretería Central", "Supermercado Norte"}},
		{name: "partial name ignoring case", company: companyA, filter: repository.ListFilter{Q: "FERRETERÍA", Status: &active},
			want: []string{"Ferretería Central"}},
		{name: "exact document", company: companyA, filter: repository.ListFilter{Q: "1020304050"},
			want: []string{"Ferretería Central"}},
		{name: "email", company: companyA, filter: repository.ListFilter{Q: "norte@super"},
			want: []string{"Supermercado Norte"}},
		{name: "inactive only", company: companyA, filter: repository.ListFilter{Status: &inactive},
			want: []string{"Ferretería Cerrada"}},
		{name: "all statuses newest first", company: companyA, filter: repository.ListFilter{Q: "ferre", SortBy: repository.SortByCreatedAt, SortOrder: repository.SortDesc},
			want: []string{"Ferretería Cerrada", "Ferretería Central"}},
		{name: "percent sign matches literally", company: companyA, filter: repository.ListFilter{Q: "100%"},
			want: []string{"Farmacia 100% Sur"}},
		{name: "lone wildcard does not match everything", company: companyA, filter: repository.ListFilter{Q: "_"},
			want: []string{}},
		{name: "no matches", company: companyA, filter: repository.ListFilter{Q: "no-existe"},
			want: []string{}},
		{name: "other company is isolated", company: companyB, filter: repository.ListFilter{},
			want: []string{"Ferretería Central"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customers, total, err := repo.List(ctx, tt.company, tt.filter, 0, 20)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if total != int64(len(tt.want)) || len(customers) != len(tt.want) {
				t.Fatalf("total = %d, len = %d, want %d", total, len(customers), len(tt.want))
			}
			for i, c := range customers {
				if c.LegalName != tt.want[i] || c.CompanyID != tt.company {
					t.Errorf("[%d] = %s (%s), want %s", i, c.LegalName, c.CompanyID, tt.want[i])
				}
			}
		})
	}
}

func TestListPaginatesWithStableTotal(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, _ := seedCustomers(t, db, repo)
	filter := repository.ListFilter{SortBy: repository.SortByLegalName, SortOrder: repository.SortAsc}

	first, total, err := repo.List(context.Background(), companyA, filter, 0, 2)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	second, total2, err := repo.List(context.Background(), companyA, filter, 2, 2)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 4 || total2 != 4 || len(first) != 2 || len(second) != 2 {
		t.Fatalf("totals %d/%d, pages %d/%d; want 4/4, 2/2", total, total2, len(first), len(second))
	}
	if first[1].LegalName >= second[0].LegalName {
		t.Errorf("pages overlap or are unordered: %s then %s", first[1].LegalName, second[0].LegalName)
	}
}

func TestListDoesNotModifyRows(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, _ := seedCustomers(t, db, repo)

	snapshot := func() map[uuid.UUID]string {
		var rows []model.CustomerModel
		db.DB.Unscoped().Where("company_id = ?", companyA).Find(&rows)
		result := map[uuid.UUID]string{}
		for _, r := range rows {
			stamp := "nil"
			if r.UpdatedAt != nil {
				stamp = r.UpdatedAt.UTC().Format(time.RFC3339Nano)
			}
			result[r.ID] = r.Status + "|" + stamp
		}
		return result
	}

	before := snapshot()
	for _, q := range []string{"", "ferre", "1020304050"} {
		if _, _, err := repo.List(context.Background(), companyA, repository.ListFilter{Q: q}, 0, 20); err != nil {
			t.Fatalf("List() error = %v", err)
		}
	}
	after := snapshot()

	for id, stamp := range before {
		if after[id] != stamp {
			t.Errorf("customer %s changed: %s -> %s", id, stamp, after[id])
		}
	}
}
