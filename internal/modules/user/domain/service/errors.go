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

	// ErrPermissionNotFound is returned when the permission id does not exist
	// (PATCH/DELETE /permissions/:id).
	ErrPermissionNotFound = sharedErrors.New(sharedErrors.CodeNotFound,
		"permission not found", 404)

	// ErrPermissionSystemProtected rejects deleting a permission of the seed.
	ErrPermissionSystemProtected = sharedErrors.New(sharedErrors.CodeConflict,
		"system permission cannot be deleted", 409).
		WithDetails(map[string]string{"system": "system permission cannot be deleted"})

	// ErrPermissionAssignedToRole rejects deleting a permission still assigned
	// to one or more roles.
	ErrPermissionAssignedToRole = sharedErrors.New(sharedErrors.CodeConflict,
		"permission is assigned to a role", 409).
		WithDetails(map[string]string{"roles": "permission is assigned to one or more roles"})

	// ErrRoleNotFound is returned when the role id does not exist or is not
	// visible to the caller (HU-082-04).
	ErrRoleNotFound = sharedErrors.New(sharedErrors.CodeNotFound,
		"role not found", 404)

	// ErrRoleSystemProtected rejects deactivating a critical system role, such
	// as superadmin, whose availability cannot be disabled (HU-082-04).
	ErrRoleSystemProtected = sharedErrors.New(sharedErrors.CodeConflict,
		"system role cannot be deactivated", 409).
		WithDetails(map[string]string{"system": "system role cannot be deactivated"})

	// ErrRoleInactive rejects assigning a role that is deactivated: an inactive
	// role can no longer be used for new assignments or authorizations
	// (HU-082-04).
	ErrRoleInactive = sharedErrors.New(sharedErrors.CodeValidation,
		"role is inactive and cannot be assigned", 400).
		WithDetails(map[string]string{"role": "role is inactive"})
)
