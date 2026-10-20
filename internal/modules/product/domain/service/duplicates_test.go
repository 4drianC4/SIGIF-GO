package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// conflicts returns the duplicated values of a 409, failing when err is not one.
func conflicts(t *testing.T, err error) []service.DuplicateConflict {
	t.Helper()
	appErr := assertAppError(t, err, sharedErrors.CodeConflict, 409)
	details, ok := appErr.Details.(service.DuplicateDetails)
	if !ok {
		t.Fatalf("details = %#v, want DuplicateDetails", appErr.Details)
	}
	return details.Conflicts
}

func fields(list []service.DuplicateConflict) []string {
	out := make([]string, len(list))
	for i, c := range list {
		out[i] = c.Field
	}
	return out
}

// barcodeProduct registers a product with a SKU and a barcode in its own
// category and returns it.
func (f fixture) barcodeProduct(t *testing.T, sku, barcode string) *entity.Product {
	t.Helper()
	return f.product(t, func(p *service.CreateProductParams) {
		p.Name = "Producto " + sku
		p.SKU = sku
		p.Barcode = barcode
	})
}

func TestCreateProductReportsTheProductHoldingTheSKU(t *testing.T) {
	f := newFixture()
	existing := f.barcodeProduct(t, "ABA-412", "7750000000001")

	params := validProduct(companyA, existing.CategoryID)
	params.Name, params.SKU = "Otro", "ABA-412"
	_, err := f.svc.CreateProduct(context.Background(), params)

	list := conflicts(t, err)
	if len(list) != 1 || list[0].Field != "sku" || list[0].Code != service.CodeSKUDuplicate || list[0].Value != "ABA-412" {
		t.Fatalf("conflicts = %+v", list)
	}
	holder := list[0].ExistingProduct
	if holder.ID != existing.ID.String() || holder.Name != existing.Name || holder.Deleted || holder.Status != entity.StatusActive {
		t.Errorf("existing product = %+v, want %s", holder, existing.Name)
	}
	if f.store.ProductCount() != 1 {
		t.Errorf("a product was stored despite the conflict")
	}
}

func TestCreateProductReportsSKUAndBarcodeTogether(t *testing.T) {
	f := newFixture()
	skuHolder := f.barcodeProduct(t, "ABA-412", "7750000000001")
	barcodeHolder := f.barcodeProduct(t, "ABA-500", "7750000000002")

	params := validProduct(companyA, skuHolder.CategoryID)
	params.Name, params.SKU, params.Barcode = "Otro", "aba-412", "7750000000002"
	_, err := f.svc.CreateProduct(context.Background(), params)

	list := conflicts(t, err)
	if got := fields(list); len(got) != 2 || got[0] != "sku" || got[1] != "barcode" {
		t.Fatalf("fields = %v, want [sku barcode]", got)
	}
	if list[0].ExistingProduct.ID != skuHolder.ID.String() || list[1].ExistingProduct.ID != barcodeHolder.ID.String() {
		t.Errorf("each conflict must name its own product: %+v", list)
	}
	if list[1].Code != service.CodeBarcodeDuplicate {
		t.Errorf("barcode code = %s", list[1].Code)
	}
	if f.store.ProductCount() != 2 {
		t.Errorf("a product was stored despite the conflicts")
	}
}

func TestSKUComparisonIgnoresCaseAndOuterSpaces(t *testing.T) {
	f := newFixture()
	existing := f.barcodeProduct(t, "ABC-01", "")

	for _, sku := range []string{"abc-01", "  ABC-01  ", " Abc-01"} {
		params := validProduct(companyA, existing.CategoryID)
		params.Name, params.SKU = "Otro "+sku, sku
		_, err := f.svc.CreateProduct(context.Background(), params)
		if list := conflicts(t, err); list[0].Value != "ABC-01" {
			t.Errorf("SKU %q: conflict value = %q, want the normalized ABC-01", sku, list[0].Value)
		}
	}
}

func TestUniquenessIsPerCompany(t *testing.T) {
	f := newFixture()
	f.barcodeProduct(t, "ABA-412", "7750000000001")

	params := validProduct(companyB, f.category(t, companyB, "Abarrotes", nil).ID)
	params.SKU, params.Barcode = "ABA-412", "7750000000001"
	if _, err := f.svc.CreateProduct(context.Background(), params); err != nil {
		t.Fatalf("another company may use the same SKU and barcode: %v", err)
	}
}

func TestProductsWithoutBarcodeOrSKUAreNotDuplicates(t *testing.T) {
	f := newFixture()
	for i := range 3 {
		f.product(t, func(p *service.CreateProductParams) {
			p.Name = "Sin códigos " + string(rune('A'+i))
			p.SKU, p.Barcode = "", ""
		})
	}
	if f.store.ProductCount() != 3 {
		t.Errorf("products without SKU and barcode = %d, want 3", f.store.ProductCount())
	}
}

func TestInactiveAndDeletedProductsKeepTheirSKUAndBarcode(t *testing.T) {
	f := newFixture()
	inactive := f.barcodeProduct(t, "INA-1", "7750000000001")
	deleted := f.barcodeProduct(t, "DEL-1", "7750000000002")
	f.store.Products[inactive.ID].Status = entity.StatusInactive
	if err := f.svc.DeleteProduct(context.Background(), service.DeleteProductParams{CompanyID: companyA, ProductID: deleted.ID}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, sku, barcode string
		holder             *entity.Product
		wantDeleted        bool
		wantStatus         entity.Status
	}{
		{"SKU of an inactive product", "INA-1", "", inactive, false, entity.StatusInactive},
		{"barcode of an inactive product", "", "7750000000001", inactive, false, entity.StatusInactive},
		{"SKU of a deleted product", "del-1", "", deleted, true, entity.StatusActive},
		{"barcode of a deleted product", "", "7750000000002", deleted, true, entity.StatusActive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := validProduct(companyA, inactive.CategoryID)
			params.Name, params.SKU, params.Barcode = "Nuevo", tt.sku, tt.barcode
			_, err := f.svc.CreateProduct(context.Background(), params)
			holder := conflicts(t, err)[0].ExistingProduct
			if holder.ID != tt.holder.ID.String() || holder.Deleted != tt.wantDeleted || holder.Status != tt.wantStatus {
				t.Errorf("existing product = %+v", holder)
			}
		})
	}
}

func TestUpdateProductReportsEveryConflictButNotItself(t *testing.T) {
	f := newFixture()
	f.barcodeProduct(t, "OTRO-1", "7750000000001")
	product := f.barcodeProduct(t, "PROPIO-1", "7750000000002")

	update := service.UpdateProductParams{
		CompanyID: companyA, ProductID: product.ID, CategoryID: product.CategoryID, UnitOfMeasureID: product.UnitOfMeasureID,
		Name: product.Name, SKU: "propio-1", Barcode: product.Barcode, Cost: product.Cost, SalePrice: product.SalePrice,
	}
	if _, err := f.svc.UpdateProduct(context.Background(), update); err != nil {
		t.Fatalf("keeping its own SKU and barcode must be allowed: %v", err)
	}

	update.SKU, update.Barcode = "OTRO-1", "7750000000001"
	_, err := f.svc.UpdateProduct(context.Background(), update)
	if got := fields(conflicts(t, err)); len(got) != 2 {
		t.Fatalf("fields = %v, want sku and barcode", got)
	}
	if stored := f.store.Products[product.ID]; stored.SKU != "PROPIO-1" {
		t.Errorf("the failed update changed the product: %+v", stored)
	}
}

func TestValidateDuplicateListsEveryConflict(t *testing.T) {
	f := newFixture()
	existing := f.barcodeProduct(t, "ABA-412", "7750000000001")

	result, err := f.svc.ValidateDuplicate(context.Background(), service.ValidateDuplicateParams{
		CompanyID: companyA, SKU: " aba-412 ", Barcode: " 7750000000001 ",
	})
	if err != nil || !result.Exists || result.Field != "sku" || len(result.Conflicts) != 2 {
		t.Fatalf("result = %+v, %v", result, err)
	}
	if result.Conflicts[1].ExistingProduct.ID != existing.ID.String() {
		t.Errorf("conflicts = %+v", result.Conflicts)
	}

	result, err = f.svc.ValidateDuplicate(context.Background(), service.ValidateDuplicateParams{CompanyID: companyA, SKU: "NUEVO-1"})
	if err != nil || result.Exists || result.Field != "" || len(result.Conflicts) != 0 {
		t.Fatalf("free SKU: %+v, %v", result, err)
	}
}

// racingProducts lets a competitor store the same SKU between the checks and
// the insert, as a concurrent request would; the insert then hits the unique
// index.
type racingProducts struct {
	repository.ProductRepository
	competitor *entity.Product
}

func (r racingProducts) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (r racingProducts) Create(ctx context.Context, product *entity.Product) error {
	if err := r.ProductRepository.Create(ctx, r.competitor); err != nil {
		return err
	}
	return service.ErrSKUTaken
}

func TestAUniqueIndexViolationIsReportedLikeTheCheck(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Abarrotes", nil)
	competitor := &entity.Product{
		ID: uuid.New(), CompanyID: companyA, CategoryID: category.ID, UnitOfMeasureID: validProduct(companyA, category.ID).UnitOfMeasureID,
		Name: "Ganador", SKU: "ABA-412", Status: entity.StatusActive,
	}
	svc := service.NewCatalogService(racingProducts{f.store.ProductRepository(), competitor}, f.store.CategoryRepository(),
		f.store.UnitRepository(), f.store.TaxRepository(), f.store.PriceHistoryRepository(), f.store.StockRepository(), clock.NewMockClock(now))

	_, err := svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
	list := conflicts(t, err)
	if list[0].Field != "sku" || list[0].ExistingProduct.ID != competitor.ID.String() || !errors.Is(err, service.ErrSKUTaken) {
		t.Fatalf("conflicts = %+v, err = %v", list, err)
	}
}
