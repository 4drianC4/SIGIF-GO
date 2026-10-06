package service

import (
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var (
	ErrCustomerNotFound = sharedErrors.New(sharedErrors.CodeNotFound, "customer not found", 404)
	ErrDocumentTaken    = sharedErrors.New(sharedErrors.CodeConflict,
		"a customer with this document already exists in this company", 409).
		WithDetails(map[string]string{"document_number": "already registered for another customer"})
)
