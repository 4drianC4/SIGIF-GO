package service

import (
	"github.com/sigif/sigif-go/internal/shared/errors"
)

var (
	ErrCompanyRequired   = errors.New(errors.CodeBadRequest, "company is required", 400)
	ErrProductRequired   = errors.New(errors.CodeBadRequest, "product is required", 400)
	ErrProductNotFound   = errors.New(errors.CodeNotFound, "product not found", 404)
	ErrInsufficientStock = errors.New(errors.CodeConflict, "insufficient stock", 409)
)
