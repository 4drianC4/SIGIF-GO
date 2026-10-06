//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	gormlib "gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
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

func newCustomer(companyID uuid.UUID, name, doc string) *entity.Customer {
	clk := clock.NewMockClock(time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC))
	var docPtr *string
	if doc != "" {
		docPtr = strPtr(doc)
	}
	return entity.NewCustomer(clk, companyID, name, entity.DocumentTypeTaxID, docPtr, nil, nil)
}

func countRows(t *testing.T, db *sharedDatabase.Database, companyIDs ...uuid.UUID) int64 {
	t.Helper()
	var count int64
	db.DB.Unscoped().Model(&model.CustomerModel{}).Where("company_id IN ?", companyIDs).Count(&count)
	return count
}

func TestUpdateNeverInserts(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, companyB := seedCustomers(t, db, repo)
	ctx := context.Background()
	before := countRows(t, db, companyA, companyB)

	existing, _, err := repo.List(ctx, companyA, repository.ListFilter{Q: "Supermercado"}, 0, 1)
	if err != nil || len(existing) != 1 {
		t.Fatalf("seed lookup: %v %v", existing, err)
	}
	deleted := newCustomer(companyA, "Borrado", "")
	if err := repo.Create(ctx, deleted); err != nil {
		t.Fatalf("create: %v", err)
	}
	deleted.SoftDelete(clock.NewRealClock())
	if err := repo.Update(ctx, deleted); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	before++

	missing := newCustomer(companyA, "No existe", "")
	foreign := *existing[0]
	foreign.CompanyID = companyB
	deleted.LegalName = "Revivido"

	for name, c := range map[string]*entity.Customer{"missing": missing, "foreign company": &foreign, "deleted": deleted} {
		t.Run(name, func(t *testing.T) {
			if err := repo.Update(ctx, c); !errors.Is(err, service.ErrCustomerNotFound) {
				t.Fatalf("Update() error = %v, want ErrCustomerNotFound", err)
			}
		})
	}

	if after := countRows(t, db, companyA, companyB); after != before {
		t.Errorf("rows = %d, want %d", after, before)
	}
}

func TestUpdateKeepsKeysAndCreatedAt(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, _ := seedCustomers(t, db, repo)
	ctx := context.Background()

	found, _, _ := repo.List(ctx, companyA, repository.ListFilter{Q: "Supermercado"}, 0, 1)
	customer := found[0]
	createdAt := customer.CreatedAt
	customer.Update(clock.NewRealClock(), "Supermercado Norte SRL", customer.DocumentType, customer.DocumentNumber, strPtr("+59170000000"), nil, nil)
	customer.CreatedAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	if err := repo.Update(ctx, customer); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	stored, err := repo.GetByID(ctx, companyA, customer.ID)
	if err != nil || stored == nil {
		t.Fatalf("GetByID() = %v, %v", stored, err)
	}
	if stored.LegalName != "Supermercado Norte SRL" || stored.Email != nil || stored.Phone == nil {
		t.Errorf("stored = %+v", stored)
	}
	if !stored.CreatedAt.Equal(createdAt) || stored.CompanyID != companyA || stored.UpdatedAt == nil {
		t.Errorf("created_at/company_id must be kept and updated_at set: %+v", stored)
	}
}

func TestWithinTransactionRollsBack(t *testing.T) {
	db := openDatabase(t)
	repo := customerGorm.NewCustomerGormRepository(db)
	companyA, _ := seedCustomers(t, db, repo)
	ctx := context.Background()
	found, _, _ := repo.List(ctx, companyA, repository.ListFilter{Q: "Supermercado"}, 0, 1)
	customer := found[0]
	boom := errors.New("boom")

	err := repo.WithinTransaction(ctx, func(ctx context.Context) error {
		customer.LegalName = "No debe persistir"
		if err := repo.Update(ctx, customer); err != nil {
			return err
		}
		return boom
	})

	if !errors.Is(err, boom) {
		t.Fatalf("WithinTransaction() error = %v, want boom", err)
	}
	stored, _ := repo.GetByID(ctx, companyA, customer.ID)
	if stored.LegalName != "Supermercado Norte" {
		t.Errorf("legal_name = %s, the update must be rolled back", stored.LegalName)
	}
}
