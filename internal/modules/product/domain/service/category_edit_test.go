package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/testutil"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func (f fixture) deleteCategory(companyID, categoryID uuid.UUID) error {
	return f.svc.DeleteCategory(context.Background(), companyID, categoryID)
}

func TestUpdateCategoryChangesOnlyWhatWasSent(t *testing.T) {
	f := newFixture()
	created, err := f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{
		CompanyID: companyA, Name: "Bebidas", Description: "Gaseosas", DefaultTax: ptr(testutil.IVAGeneral), TargetMargin: ptr(dec("25")),
	})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Category.ID
	f.product(t, func(p *service.CreateProductParams) { p.CategoryID = id })

	item, err := f.svc.UpdateCategory(context.Background(), service.UpdateCategoryParams{CompanyID: companyA, CategoryID: id, Name: ptr("  Aguas  y jugos ")})
	if err != nil {
		t.Fatal(err)
	}
	c := item.Category
	if c.Name != "Aguas y jugos" || c.Description != "Gaseosas" || c.TargetMargin == nil || !c.TargetMargin.Equal(dec("25")) {
		t.Errorf("category = %+v", c)
	}
	if item.DefaultTaxName == nil || *item.DefaultTaxName != testutil.IVAGeneral {
		t.Errorf("default tax = %v, want it kept", item.DefaultTaxName)
	}
	if item.ProductsCount != 1 {
		t.Errorf("products count = %d, want 1", item.ProductsCount)
	}

	item, err = f.svc.UpdateCategory(context.Background(), service.UpdateCategoryParams{
		CompanyID: companyA, CategoryID: id, Description: ptr(""),
		DefaultTax: entity.Some("Exempt"), TargetMargin: entity.Some(dec("0")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.Category.Description != "" || item.DefaultTaxName == nil || *item.DefaultTaxName != "Exempt" ||
		item.Category.TargetMargin == nil || !item.Category.TargetMargin.IsZero() || item.Category.Name != "Aguas y jugos" {
		t.Errorf("category = %+v, tax %v", item.Category, item.DefaultTaxName)
	}

	item, err = f.svc.UpdateCategory(context.Background(), service.UpdateCategoryParams{
		CompanyID: companyA, CategoryID: id, DefaultTax: entity.Cleared[string](), TargetMargin: entity.Cleared[decimal.Decimal](),
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.DefaultTaxName != nil || item.Category.DefaultTaxID != nil || item.Category.TargetMargin != nil {
		t.Errorf("tax and margin must be cleared: %+v", item)
	}
}

func TestUpdateCategoryWorksOnSubcategoriesAndKeepsTheParent(t *testing.T) {
	f := newFixture()
	parent := f.category(t, companyA, "Abarrotes", nil)
	child := f.category(t, companyA, "Aceites", &parent.ID)

	item, err := f.svc.UpdateCategory(context.Background(), service.UpdateCategoryParams{CompanyID: companyA, CategoryID: child.ID, Name: ptr("Aceites vegetales")})
	if err != nil {
		t.Fatal(err)
	}
	if item.Category.Name != "Aceites vegetales" || item.Category.ParentID == nil || *item.Category.ParentID != parent.ID {
		t.Errorf("subcategory = %+v", item.Category)
	}
}

func TestUpdateCategoryNameIsUniquePerLevelButNotAgainstItself(t *testing.T) {
	f := newFixture()
	abarrotes := f.category(t, companyA, "Abarrotes", nil)
	limpieza := f.category(t, companyA, "Limpieza", nil)
	aceites := f.category(t, companyA, "Aceites", &abarrotes.ID)
	f.category(t, companyA, "Vinagres", &abarrotes.ID)
	update := func(id uuid.UUID, name string) error {
		_, err := f.svc.UpdateCategory(context.Background(), service.UpdateCategoryParams{CompanyID: companyA, CategoryID: id, Name: ptr(name)})
		return err
	}

	assertAppError(t, update(limpieza.ID, "ABARROTES"), sharedErrors.CodeConflict, 409)
	assertAppError(t, update(aceites.ID, "vinagres"), sharedErrors.CodeConflict, 409)
	if err := update(abarrotes.ID, "ABARROTES"); err != nil {
		t.Errorf("changing only the case of its own name must be allowed: %v", err)
	}
	if err := update(limpieza.ID, "Limpieza"); err != nil {
		t.Errorf("sending its own name must be allowed: %v", err)
	}
	if err := update(aceites.ID, "Limpieza"); err != nil {
		t.Errorf("the same name under another parent must be allowed: %v", err)
	}
	if err := update(limpieza.ID, "Aceites"); err != nil {
		t.Errorf("a root category may share a name with a subcategory: %v", err)
	}
	if stored, _ := f.store.CategoryRepository().GetByID(context.Background(), companyA, abarrotes.ID); stored.Name != "ABARROTES" {
		t.Errorf("stored name = %q", stored.Name)
	}
}

func TestUpdateCategoryErrors(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Abarrotes", nil)
	other := f.category(t, companyA, "Limpieza", nil)

	tests := []struct {
		name   string
		params service.UpdateCategoryParams
		code   sharedErrors.ErrorCode
		status int
	}{
		{"unknown category", service.UpdateCategoryParams{CompanyID: companyA, CategoryID: uuid.New(), Name: ptr("X1")}, sharedErrors.CodeNotFound, 404},
		{"category of another company", service.UpdateCategoryParams{CompanyID: companyB, CategoryID: category.ID, Name: ptr("X1")}, sharedErrors.CodeNotFound, 404},
		{"without company", service.UpdateCategoryParams{CategoryID: category.ID, Name: ptr("X1")}, sharedErrors.CodeBadRequest, 400},
		{"unknown tax", service.UpdateCategoryParams{CompanyID: companyA, CategoryID: category.ID, DefaultTax: entity.Some("No existe")}, sharedErrors.CodeNotFound, 404},
		{"margin above 100", service.UpdateCategoryParams{CompanyID: companyA, CategoryID: category.ID, TargetMargin: entity.Some(dec("100.5"))}, sharedErrors.CodeValidation, 400},
		{"repeated name", service.UpdateCategoryParams{CompanyID: companyA, CategoryID: category.ID, Name: ptr(other.Name)}, sharedErrors.CodeConflict, 409},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.svc.UpdateCategory(context.Background(), tt.params)
			assertAppError(t, err, tt.code, tt.status)
			if stored, _ := f.store.CategoryRepository().GetByID(context.Background(), companyA, category.ID); stored.Name != "Abarrotes" || stored.TargetMargin != nil {
				t.Errorf("the category changed despite the error: %+v", stored)
			}
		})
	}
}

func TestUpdateCategoryIsAllowedWhileInactive(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Abarrotes", nil)
	f.store.Categories[category.ID].Status = entity.StatusInactive
	if _, err := f.svc.UpdateCategory(context.Background(), service.UpdateCategoryParams{CompanyID: companyA, CategoryID: category.ID, Name: ptr("Almacén")}); err != nil {
		t.Fatalf("an inactive category can be edited: %v", err)
	}
}

func TestDeleteCategory(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas", nil)

	assertAppError(t, f.deleteCategory(companyB, category.ID), sharedErrors.CodeNotFound, 404)
	assertAppError(t, f.deleteCategory(companyA, uuid.New()), sharedErrors.CodeNotFound, 404)
	assertAppError(t, f.deleteCategory(uuid.Nil, category.ID), sharedErrors.CodeBadRequest, 400)
	if got, _ := f.store.CategoryRepository().GetByID(context.Background(), companyA, category.ID); got == nil {
		t.Fatal("a failed delete removed the category")
	}

	if err := f.deleteCategory(companyA, category.ID); err != nil {
		t.Fatalf("DeleteCategory() error = %v", err)
	}
	if got, _ := f.store.CategoryRepository().GetByID(context.Background(), companyA, category.ID); got != nil {
		t.Error("the deleted category can still be read")
	}
	assertAppError(t, f.deleteCategory(companyA, category.ID), sharedErrors.CodeNotFound, 404)
	if _, err := f.svc.UpdateCategory(context.Background(), service.UpdateCategoryParams{CompanyID: companyA, CategoryID: category.ID, Name: ptr("X1")}); !errors.Is(err, service.ErrCategoryNotFound) {
		t.Errorf("editing a deleted category: error = %v, want ErrCategoryNotFound", err)
	}
	list, _ := f.svc.ListCategories(context.Background(), companyA)
	if len(list) != 0 {
		t.Errorf("the deleted category is still listed: %d", len(list))
	}

	f.category(t, companyA, "Bebidas", nil) // its name is free again
}

func TestDeleteCategoryIsRefusedWhileItHasSubcategories(t *testing.T) {
	f := newFixture()
	parent := f.category(t, companyA, "Abarrotes", nil)
	child := f.category(t, companyA, "Aceites", &parent.ID)

	err := f.deleteCategory(companyA, parent.ID)
	assertAppError(t, err, sharedErrors.CodeConflict, 409)
	if !errors.Is(err, service.ErrCategoryHasSubcategories) {
		t.Errorf("error = %v, want ErrCategoryHasSubcategories", err)
	}

	f.store.Categories[child.ID].Status = entity.StatusInactive
	if !errors.Is(f.deleteCategory(companyA, parent.ID), service.ErrCategoryHasSubcategories) {
		t.Error("an inactive subcategory must still block the delete")
	}

	if err := f.deleteCategory(companyA, child.ID); err != nil {
		t.Fatalf("deleting the subcategory: %v", err)
	}
	if err := f.deleteCategory(companyA, parent.ID); err != nil {
		t.Errorf("a deleted subcategory must not block: %v", err)
	}
}

func TestDeleteCategoryIsRefusedWhileItHasProducts(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas", nil)
	product := f.product(t, func(p *service.CreateProductParams) { p.CategoryID = category.ID })

	err := f.deleteCategory(companyA, category.ID)
	assertAppError(t, err, sharedErrors.CodeConflict, 409)
	if !errors.Is(err, service.ErrCategoryHasProducts) {
		t.Errorf("error = %v, want ErrCategoryHasProducts", err)
	}

	f.store.Products[product.ID].Status = entity.StatusInactive
	if !errors.Is(f.deleteCategory(companyA, category.ID), service.ErrCategoryHasProducts) {
		t.Error("an inactive product must still block the delete")
	}

	if err := f.svc.DeleteProduct(context.Background(), service.DeleteProductParams{CompanyID: companyA, ProductID: product.ID}); err != nil {
		t.Fatal(err)
	}
	if err := f.deleteCategory(companyA, category.ID); err != nil {
		t.Errorf("a deleted product must not block: %v", err)
	}
}

func TestProductsAndSubcategoriesCannotBeCreatedInADeletedCategory(t *testing.T) {
	f := newFixture()
	category := f.category(t, companyA, "Bebidas", nil)
	if err := f.deleteCategory(companyA, category.ID); err != nil {
		t.Fatal(err)
	}

	_, err := f.svc.CreateProduct(context.Background(), validProduct(companyA, category.ID))
	if !errors.Is(err, service.ErrCategoryNotFound) {
		t.Errorf("product in a deleted category: error = %v, want ErrCategoryNotFound", err)
	}
	_, err = f.svc.CreateCategory(context.Background(), service.CreateCategoryParams{CompanyID: companyA, Name: "Gaseosas", ParentID: &category.ID})
	if !errors.Is(err, service.ErrParentNotFound) {
		t.Errorf("subcategory of a deleted category: error = %v, want ErrParentNotFound", err)
	}
}
