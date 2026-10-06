package service

import (
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var (
	ErrCompanyRequired      = sharedErrors.New(sharedErrors.CodeBadRequest, "company is required", 400)
	ErrCategoryNameTaken    = sharedErrors.New(sharedErrors.CodeConflict, "category name already exists", 409)
	ErrProductNameTaken     = sharedErrors.New(sharedErrors.CodeConflict, "product name already exists", 409)
	ErrSKUTaken             = sharedErrors.New(sharedErrors.CodeConflict, "SKU already exists", 409)
	ErrBarcodeTaken         = sharedErrors.New(sharedErrors.CodeConflict, "barcode already exists", 409)
	ErrCategoryNotFound     = sharedErrors.New(sharedErrors.CodeNotFound, "category not found", 404)
	ErrParentNotFound       = sharedErrors.New(sharedErrors.CodeNotFound, "parent category not found", 404)
	ErrCategoryInactive     = sharedErrors.New(sharedErrors.CodeBadRequest, "category is inactive", 400)
	ErrUnitNotFound         = sharedErrors.New(sharedErrors.CodeNotFound, "unit of measure not found", 404)
	ErrTaxNotFound          = sharedErrors.New(sharedErrors.CodeNotFound, "tax not found", 404)
	ErrDuplicateCheckFields = sharedErrors.New(sharedErrors.CodeBadRequest, "name, sku or barcode is required", 400)
)
