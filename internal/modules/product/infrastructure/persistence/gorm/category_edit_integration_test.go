//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	productGorm "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/product/testutil"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func ptr[T any](v T) *T { return &v }

func newCategoryService(r repos) *service.CatalogService {
	return service.NewCatalogService(r.products, r.categories, r.units, r.taxes,
		productGorm.NewPriceHistoryGormRepository(r.db), productGorm.NewStockGormRepository(r.db), clock.NewRealClock())
}

func TestUpdateAndDeleteCategoryWithPostgres(t *testing.T) {
	r := openDatabase(t)
	svc := newCategoryService(r)
	ctx := context.Background()
	companyID := uuid.New()
	create := func(name string, parent *uuid.UUID) uuid.UUID {
		created, err := svc.CreateCategory(ctx, service.CreateCategoryParams{
			CompanyID: companyID, Name: name, ParentID: parent, DefaultTax: ptr(testutil.IVAGeneral), TargetMargin: ptr(decimal.NewFromInt(25)),
		})
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		return created.Category.ID
	}
	root := create("Abarrotes", nil)
	other := create("Limpieza", nil)
	child := create("Aceites", &root)

	// Update keeps what was not sent and clears what was sent as null.
	name := "Almacén"
	item, err := svc.UpdateCategory(ctx, service.UpdateCategoryParams{CompanyID: companyID, CategoryID: root, Name: &name})
	if err != nil || item.Category.Name != "Almacén" || item.DefaultTaxName == nil || item.Category.TargetMargin == nil {
		t.Fatalf("rename: %+v, %v", item, err)
	}
	cleared, err := svc.UpdateCategory(ctx, service.UpdateCategoryParams{CompanyID: companyID, CategoryID: root, TargetMargin: entity.Cleared[decimal.Decimal]()})
	if err != nil || cleared.Category.TargetMargin != nil || cleared.Category.Name != "Almacén" {
		t.Fatalf("clear margin: %+v, %v", cleared, err)
	}
	var stored struct {
		TargetMargin *decimal.Decimal
		Name         string
	}
	r.db.DB.Table("categories").Select("target_margin, name").Where("id = ?", root).Scan(&stored)
	if stored.TargetMargin != nil || stored.Name != "Almacén" {
		t.Errorf("stored row = %+v", stored)
	}

	// The name is unique per level ignoring case, but not against itself.
	clash := "LIMPIEZA"
	if _, err := svc.UpdateCategory(ctx, service.UpdateCategoryParams{CompanyID: companyID, CategoryID: root, Name: &clash}); !errors.Is(err, service.ErrCategoryNameTaken) {
		t.Errorf("repeated name: error = %v, want ErrCategoryNameTaken", err)
	}
	same := "LIMPIEZA"
	if _, err := svc.UpdateCategory(ctx, service.UpdateCategoryParams{CompanyID: companyID, CategoryID: other, Name: &same}); err != nil {
		t.Errorf("changing the case of its own name: %v", err)
	}
	// The index is the safety net when the check is skipped, also for a name
	// that only differs in case.
	rootCategory, err := r.categories.GetByID(ctx, companyID, root)
	if err != nil || rootCategory == nil {
		t.Fatalf("reload root: %v", err)
	}
	dup := *rootCategory
	dup.Name = "limpieza"
	if err := r.categories.Update(ctx, &dup); !errors.Is(err, service.ErrCategoryNameTaken) {
		t.Errorf("unique index on update: error = %v, want ErrCategoryNameTaken", err)
	}

	// Subcategories and products block the delete; counts include inactive rows.
	if err := svc.DeleteCategory(ctx, companyID, root); !errors.Is(err, service.ErrCategoryHasSubcategories) {
		t.Errorf("parent with a subcategory: error = %v", err)
	}
	units, _ := r.units.List(ctx)
	product, err := svc.CreateProduct(ctx, productParams(companyID, child, units[0].ID, "Aceite", "ACE-1", ""))
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteCategory(ctx, companyID, child); !errors.Is(err, service.ErrCategoryHasProducts) {
		t.Errorf("category with a product: error = %v", err)
	}
	r.db.DB.Exec(`UPDATE products SET status = 'inactive' WHERE id = ?`, product.Product.ID)
	if err := svc.DeleteCategory(ctx, companyID, child); !errors.Is(err, service.ErrCategoryHasProducts) {
		t.Errorf("category with an inactive product: error = %v", err)
	}
	if err := svc.DeleteProduct(ctx, service.DeleteProductParams{CompanyID: companyID, ProductID: product.Product.ID}); err != nil {
		t.Fatal(err)
	}

	// Delete is logical and frees the name.
	if err := svc.DeleteCategory(ctx, companyID, child); err != nil {
		t.Fatalf("delete subcategory: %v", err)
	}
	if err := svc.DeleteCategory(ctx, companyID, root); err != nil {
		t.Fatalf("delete parent: %v", err)
	}
	var rows, live int64
	r.db.DB.Table("categories").Where("id IN ?", []uuid.UUID{root, child}).Count(&rows)
	r.db.DB.Table("categories").Where("id IN ? AND deleted_at IS NULL", []uuid.UUID{root, child}).Count(&live)
	if rows != 2 || live != 0 {
		t.Errorf("rows = %d, live = %d; want the rows kept and marked as deleted", rows, live)
	}
	if err := svc.DeleteCategory(ctx, companyID, root); !errors.Is(err, service.ErrCategoryNotFound) {
		t.Errorf("deleting twice: error = %v", err)
	}
	create("Almacén", nil)
	create("Aceites", nil)
	list, err := svc.ListCategories(ctx, companyID)
	if err != nil || len(list) != 3 {
		t.Errorf("listing = %d categories, %v; want 3", len(list), err)
	}
}

// Deleting a category while a product or a subcategory is created in it must
// never leave a live row under a deleted category: either the delete is
// refused or the creation finds the category gone.
func TestDeleteCategoryRacesWithCreationsInIt(t *testing.T) {
	r := openDatabase(t)
	svc := newCategoryService(r)
	ctx := context.Background()
	companyID := uuid.New()
	units, _ := r.units.List(ctx)

	const rounds = 40
	var wg sync.WaitGroup
	var mu sync.Mutex
	refusedDeletes, deletedCategories := 0, 0
	for i := range rounds {
		created, err := svc.CreateCategory(ctx, service.CreateCategoryParams{CompanyID: companyID, Name: "Ronda " + uuid.NewString()[:8]})
		if err != nil {
			t.Fatal(err)
		}
		id := created.Category.ID
		wg.Add(3)
		go func() {
			defer wg.Done()
			_, _ = svc.CreateProduct(ctx, productParams(companyID, id, units[0].ID, "Producto "+uuid.NewString()[:8], "R"+uuid.NewString()[:8], ""))
		}()
		go func() {
			defer wg.Done()
			_, _ = svc.CreateCategory(ctx, service.CreateCategoryParams{CompanyID: companyID, Name: "Hija " + uuid.NewString()[:8], ParentID: &id})
		}()
		go func() {
			defer wg.Done()
			err := svc.DeleteCategory(ctx, companyID, id)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				deletedCategories++
			case errors.Is(err, service.ErrCategoryHasProducts), errors.Is(err, service.ErrCategoryHasSubcategories):
				refusedDeletes++
			default:
				t.Errorf("round %d: unexpected delete error: %v", i, err)
			}
		}()
	}
	wg.Wait()

	var orphanProducts, orphanCategories int64
	r.db.DB.Raw(`SELECT count(*) FROM products p JOIN categories c ON c.id = p.category_id
		WHERE p.company_id = ? AND p.deleted_at IS NULL AND c.deleted_at IS NOT NULL`, companyID).Scan(&orphanProducts)
	r.db.DB.Raw(`SELECT count(*) FROM categories child JOIN categories parent ON parent.id = child.parent_id
		WHERE child.company_id = ? AND child.deleted_at IS NULL AND parent.deleted_at IS NOT NULL`, companyID).Scan(&orphanCategories)
	if orphanProducts != 0 || orphanCategories != 0 {
		t.Fatalf("live rows under a deleted category: %d products, %d subcategories", orphanProducts, orphanCategories)
	}
	t.Logf("%d categories deleted, %d deletes refused", deletedCategories, refusedDeletes)
}
