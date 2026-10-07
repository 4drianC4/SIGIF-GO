package service

import (
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var (
	// ErrPermissionCodeTaken rejects a code that is already registered (HU-082-01).
	ErrPermissionCodeTaken = sharedErrors.New(sharedErrors.CodeConflict,
		"permission code is already registered", 409).
		WithDetails(map[string]string{"code": "already registered"})

	ErrUnknownPermissionModule = sharedErrors.New(sharedErrors.CodeValidation,
		"unknown module", 400).
		WithDetails(map[string]string{"module": "not a supported module"})

	ErrUnknownPermissionOperation = sharedErrors.New(sharedErrors.CodeValidation,
		"unknown operation", 400).
		WithDetails(map[string]string{"operation": "not a supported operation"})

	ErrPermissionCodeMismatch = sharedErrors.New(sharedErrors.CodeValidation,
		"code must match module.operation", 400).
		WithDetails(map[string]string{"code": "must match module.operation"})
)
