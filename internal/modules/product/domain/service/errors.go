package service

import (
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var (
	ErrCategoryNameTaken = sharedErrors.New(sharedErrors.CodeConflict, "a category with this name already exists", 409)
	ErrSKUTaken          = sharedErrors.New(sharedErrors.CodeConflict, "a product with this sku already exists", 409)
	ErrBarcodeTaken      = sharedErrors.New(sharedErrors.CodeConflict, "a product with this barcode already exists", 409)
	ErrCategoryNotFound  = sharedErrors.New(sharedErrors.CodeNotFound, "category not found", 404)
	ErrCategoryInactive  = sharedErrors.New(sharedErrors.CodeBadRequest, "category is inactive", 400)
)
