package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// ChangeRoleStatus activates or deactivates a role (HU-082-04). Deactivating a
// role does not delete it nor any of its permission assignments: it only stops
// the role from being used for new assignments or authorizations. Roles that
// the caller cannot see (another company's custom roles) are reported as not
// found, and protected system roles cannot be deactivated.
//
// Changing a role to the status it already has is idempotent: the role is
// returned without writing to the database.
func (s *UserService) ChangeRoleStatus(
	ctx context.Context,
	companyID *uuid.UUID,
	id uuid.UUID,
	newStatus entity.RoleStatus,
) (*entity.Role, error) {
	if !newStatus.IsValid() {
		return nil, sharedErrors.New(sharedErrors.CodeValidation,
			"invalid status value; accepted values: active, inactive", 400)
	}

	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil || !roleVisibleToCompany(role, companyID) {
		return nil, ErrRoleNotFound
	}

	if newStatus == entity.RoleStatusInactive && role.IsProtected() {
		return nil, ErrRoleSystemProtected
	}

	if role.Status == newStatus {
		return s.withPermissionCount(ctx, role)
	}

	if newStatus == entity.RoleStatusActive {
		role.Activate()
	} else {
		role.Deactivate()
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}
	return s.withPermissionCount(ctx, role)
}

// withPermissionCount fills the role's permissions count so the status response
// matches the shape of the listing. The assignments are never modified here.
func (s *UserService) withPermissionCount(ctx context.Context, role *entity.Role) (*entity.Role, error) {
	permissions, err := s.permRepo.ListByRole(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	role.PermissionsCount = len(permissions)
	return role, nil
}

// roleVisibleToCompany mirrors the visibility rule of the role listing: a
// global role (company_id nil) is visible to every caller, a company role only
// to a caller of that same company.
func roleVisibleToCompany(role *entity.Role, companyID *uuid.UUID) bool {
	if role.CompanyID == nil {
		return true
	}
	return companyID != nil && *role.CompanyID == *companyID
}
