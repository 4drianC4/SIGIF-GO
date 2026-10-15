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
	ErrCompanyRequired = sharedErrors.New(sharedErrors.CodeBadRequest, "company context is required", 400)
)

type SupplierService struct {
	repo     repository.SupplierRepository
	products repository.ProductSupplierRepository
	clock    clock.Clock
}

func NewSupplierService(repo repository.SupplierRepository, products repository.ProductSupplierRepository, clk clock.Clock) *SupplierService {
	return &SupplierService{repo: repo, products: products, clock: clk}
}
