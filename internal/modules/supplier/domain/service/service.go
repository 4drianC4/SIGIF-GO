package service

import (
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var (
	ErrSupplierNotFound = sharedErrors.New(sharedErrors.CodeNotFound, "supplier not found", 404)
	ErrTaxIDTaken       = sharedErrors.New(sharedErrors.CodeConflict, "a supplier with this tax ID already exists in this company", 409).
				WithDetails(map[string]string{"tax_id": "already registered"})
	ErrCompanyRequired     = sharedErrors.New(sharedErrors.CodeBadRequest, "company context is required", 400)
	ErrProductNotFound     = sharedErrors.New(sharedErrors.CodeNotFound, "product not found", 404)
	ErrAssociationNotFound = sharedErrors.New(sharedErrors.CodeNotFound, "supplier product association not found", 404)
	ErrAlreadyAssociated   = sharedErrors.New(sharedErrors.CodeConflict, "product is already associated with this supplier", 409)
	ErrInactiveSupplier    = sharedErrors.New(sharedErrors.CodeConflict, "cannot associate an inactive supplier", 409)
	ErrInactiveProduct     = sharedErrors.New(sharedErrors.CodeConflict, "cannot associate an inactive product", 409)
	ErrPreferredConflict   = sharedErrors.New(sharedErrors.CodeConflict, "product already has a preferred supplier", 409)
)

type SupplierService struct {
	repo     repository.SupplierRepository
	products repository.ProductSupplierRepository
	clock    clock.Clock
}

func NewSupplierService(repo repository.SupplierRepository, products repository.ProductSupplierRepository, clk clock.Clock) *SupplierService {
	return &SupplierService{repo: repo, products: products, clock: clk}
}
