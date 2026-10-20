//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	productGorm "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// uniquenessEnv returns a service on Postgres and an active category of a new
// company (HU-06-01).
func uniquenessEnv(t *testing.T) (repos, *service.CatalogService, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	r := openDatabase(t)
	ctx := context.Background()
	svc := service.NewCatalogService(r.products, r.categories, r.units, r.taxes,
		productGorm.NewPriceHistoryGormRepository(r.db), productGorm.NewStockGormRepository(r.db), clock.NewRealClock())
	companyID := uuid.New()
	category, err := svc.CreateCategory(ctx, service.CreateCategoryParams{CompanyID: companyID, Name: "Abarrotes"})
	if err != nil {
		t.Fatal(err)
	}
	units, err := r.units.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return r, svc, companyID, category.Category.ID, units[0].ID
}

func productParams(companyID, categoryID, unitID uuid.UUID, name, sku, barcode string) service.CreateProductParams {
	return service.CreateProductParams{
		CompanyID: companyID, CategoryID: categoryID, UnitOfMeasureID: unitID,
		Name: name, SKU: sku, Barcode: barcode,
		Cost: decimal.RequireFromString("4.20"), SalePrice: decimal.RequireFromString("6.00"),
	}
}

func TestDeletedProductsKeepTheirSKUAndBarcodeInPostgres(t *testing.T) {
	r, svc, companyID, categoryID, unitID := uniquenessEnv(t)
	ctx := context.Background()
	created, err := svc.CreateProduct(ctx, productParams(companyID, categoryID, unitID, "Descontinuado", "DEL-1", "7750000000001"))
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteProduct(ctx, service.DeleteProductParams{CompanyID: companyID, ProductID: created.Product.ID}); err != nil {
		t.Fatal(err)
	}

	_, err = svc.CreateProduct(ctx, productParams(companyID, categoryID, unitID, "Nuevo", " del-1 ", "7750000000001"))
	var appErr *sharedErrors.AppError
	if !errors.As(err, &appErr) || appErr.StatusCode != 409 {
		t.Fatalf("error = %v, want 409", err)
	}
	details := appErr.Details.(service.DuplicateDetails)
	if len(details.Conflicts) != 2 || !details.Conflicts[0].ExistingProduct.Deleted || details.Conflicts[0].ExistingProduct.ID != created.Product.ID.String() {
		t.Fatalf("conflicts = %+v", details.Conflicts)
	}

	// The database refuses it too, even without the service checks.
	clk := clock.NewMockClock(time.Now().UTC())
	direct, err := entity.NewProduct(clk, entity.NewProductParams{
		CompanyID: companyID, CategoryID: categoryID, UnitOfMeasureID: unitID, Name: "Directo", SKU: "DEL-1",
		Cost: decimal.RequireFromString("1"), SalePrice: decimal.RequireFromString("2"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.products.Create(ctx, direct); !errors.Is(err, service.ErrSKUTaken) {
		t.Errorf("unique index: error = %v, want ErrSKUTaken", err)
	}

	// The name of the deleted product is free.
	if _, err := svc.CreateProduct(ctx, productParams(companyID, categoryID, unitID, "Descontinuado", "NEW-1", "")); err != nil {
		t.Errorf("the name of a deleted product must be free: %v", err)
	}
}

// Every request passes the checks before any insert; the unique index picks a
// single winner and the others get the detailed conflict naming it.
func TestConcurrentCreationsWithTheSameSKU(t *testing.T) {
	_, svc, companyID, categoryID, unitID := uniquenessEnv(t)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var winners []uuid.UUID
	var losers []service.DuplicateConflict
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			created, err := svc.CreateProduct(context.Background(),
				productParams(companyID, categoryID, unitID, "Producto "+uuid.NewString()[:8]+string(rune('A'+i)), "RACE-1", ""))
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				winners = append(winners, created.Product.ID)
				return
			}
			var appErr *sharedErrors.AppError
			details, ok := service.DuplicateDetails{}, false
			if errors.As(err, &appErr) {
				details, ok = appErr.Details.(service.DuplicateDetails)
			}
			if !ok || len(details.Conflicts) != 1 || details.Conflicts[0].Code != service.CodeSKUDuplicate {
				t.Errorf("loser error = %v (%#v)", err, appErr)
				return
			}
			losers = append(losers, details.Conflicts[0])
		}()
	}
	wg.Wait()

	if len(winners) != 1 || len(losers) != 19 {
		t.Fatalf("winners = %d, losers = %d; want 1 and 19", len(winners), len(losers))
	}
	for _, c := range losers {
		if c.ExistingProduct.ID != winners[0].String() {
			t.Errorf("a conflict names %s, not the winner %s", c.ExistingProduct.ID, winners[0])
		}
	}
}

// An index violation that is not about the name, the SKU or the barcode must
// reach the caller as it is, not be reported as a duplicated SKU.
func TestUnrelatedUniqueViolationIsNotReportedAsSKU(t *testing.T) {
	r, svc, companyID, categoryID, unitID := uniquenessEnv(t)
	ctx := context.Background()
	created, err := svc.CreateProduct(ctx, productParams(companyID, categoryID, unitID, "Original", "ORI-1", ""))
	if err != nil {
		t.Fatal(err)
	}

	clk := clock.NewMockClock(time.Now().UTC())
	clone, err := entity.NewProduct(clk, entity.NewProductParams{
		CompanyID: companyID, CategoryID: categoryID, UnitOfMeasureID: unitID, Name: "Otro nombre", SKU: "OTRO-1",
		Cost: decimal.RequireFromString("1"), SalePrice: decimal.RequireFromString("2"),
	})
	if err != nil {
		t.Fatal(err)
	}
	clone.ID = created.Product.ID // same primary key, different name, SKU and barcode

	err = r.products.Create(ctx, clone)
	if err == nil {
		t.Fatal("a repeated primary key was accepted")
	}
	for _, domain := range []error{service.ErrSKUTaken, service.ErrBarcodeTaken, service.ErrProductNameTaken} {
		if errors.Is(err, domain) {
			t.Errorf("primary key violation reported as %v", domain)
		}
	}
}

func TestSKUIndexViolationIsReportedAsSKU(t *testing.T) {
	r, svc, companyID, categoryID, unitID := uniquenessEnv(t)
	ctx := context.Background()
	if _, err := svc.CreateProduct(ctx, productParams(companyID, categoryID, unitID, "Original", "SKU-1", "")); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewMockClock(time.Now().UTC())
	other, err := entity.NewProduct(clk, entity.NewProductParams{
		CompanyID: companyID, CategoryID: categoryID, UnitOfMeasureID: unitID, Name: "Otro nombre", SKU: "SKU-1",
		Cost: decimal.RequireFromString("1"), SalePrice: decimal.RequireFromString("2"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.products.Create(ctx, other); !errors.Is(err, service.ErrSKUTaken) {
		t.Errorf("error = %v, want ErrSKUTaken", err)
	}
}
