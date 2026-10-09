package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// grantPermRepo reports the configured permission and counts role listings, so
// the status tests can assert that permissions are still stored but not
// effective while the role is inactive.
type grantPermRepo struct {
	noopPermRepo
	granted bool
	listed  int
}

func (r *grantPermRepo) HasPermission(_ context.Context, _ uuid.UUID, _, _ string) (bool, error) {
	return r.granted, nil
}

func (r *grantPermRepo) ListByRole(_ context.Context, _ uuid.UUID) ([]*entity.Permission, error) {
	r.listed++
	return []*entity.Permission{{ID: uuid.New(), Module: "roles", Operation: "list"}}, nil
}

type statusRoleRepo struct {
	*memoryRoleRepo
	updates int
}

func (r *statusRoleRepo) Update(ctx context.Context, role *entity.Role) error {
	r.updates++
	return r.memoryRoleRepo.Update(ctx, role)
}

func newStatusService(role *entity.Role, perm *grantPermRepo) (*UserService, *statusRoleRepo) {
	roles := &statusRoleRepo{memoryRoleRepo: &memoryRoleRepo{byName: map[string]*entity.Role{}}}
	if role != nil {
		roles.byName[role.Name] = role
	}
	svc := NewUserService(newMemoryUserRepo(), roles, perm, clock.NewMockClock(time.Now()), &memoryCompanyRepo{})
	return svc, roles
}

func TestChangeRoleStatusDeactivateKeepsConfiguration(t *testing.T) {
	companyID := uuid.New()
	role := &entity.Role{ID: uuid.New(), CompanyID: &companyID, Name: "Vendedor", Status: entity.RoleStatusActive}
	perm := &grantPermRepo{granted: true}
	svc, roles := newStatusService(role, perm)
	user := activeUserWithRole(role.ID)
	if err := svc.repo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	got, err := svc.ChangeRoleStatus(context.Background(), &companyID, role.ID, entity.RoleStatusInactive)
	if err != nil {
		t.Fatalf("ChangeRoleStatus: %v", err)
	}
	if got.Status != entity.RoleStatusInactive {
		t.Fatalf("status = %q, want inactive", got.Status)
	}
	if roles.updates != 1 {
		t.Fatalf("updates = %d, want 1", roles.updates)
	}
	if stored := roles.byName["Vendedor"]; stored.Name != "Vendedor" || stored.CompanyID == nil || *stored.CompanyID != companyID {
		t.Fatalf("configuration changed: %+v", stored)
	}

	// An inactive role grants no effective permission to an active user.
	allowed, err := svc.HasPermission(context.Background(), user.ID, "roles", "list")
	if err != nil {
		t.Fatalf("HasPermission: %v", err)
	}
	if allowed {
		t.Fatal("inactive role must not grant permissions")
	}
}

func TestChangeRoleStatusReactivateRestoresAvailability(t *testing.T) {
	companyID := uuid.New()
	role := &entity.Role{ID: uuid.New(), CompanyID: &companyID, Name: "Vendedor", Status: entity.RoleStatusInactive}
	perm := &grantPermRepo{granted: true}
	svc, roles := newStatusService(role, perm)

	got, err := svc.ChangeRoleStatus(context.Background(), &companyID, role.ID, entity.RoleStatusActive)
	if err != nil {
		t.Fatalf("ChangeRoleStatus: %v", err)
	}
	if got.Status != entity.RoleStatusActive {
		t.Fatalf("status = %q, want active", got.Status)
	}
	if roles.updates != 1 {
		t.Fatalf("updates = %d, want 1", roles.updates)
	}
}

func TestChangeRoleStatusIsIdempotent(t *testing.T) {
	companyID := uuid.New()
	role := &entity.Role{ID: uuid.New(), CompanyID: &companyID, Name: "Vendedor", Status: entity.RoleStatusActive}
	svc, roles := newStatusService(role, &grantPermRepo{})

	got, err := svc.ChangeRoleStatus(context.Background(), &companyID, role.ID, entity.RoleStatusActive)
	if err != nil {
		t.Fatalf("ChangeRoleStatus: %v", err)
	}
	if got.Status != entity.RoleStatusActive {
		t.Fatalf("status = %q, want active", got.Status)
	}
	if roles.updates != 0 {
		t.Fatalf("updates = %d, want 0 for a no-op change", roles.updates)
	}
}

func TestChangeRoleStatusProtectedSuperadmin(t *testing.T) {
	role := &entity.Role{ID: uuid.New(), Name: entity.RoleSuperadmin, IsSystem: true, Status: entity.RoleStatusActive}
	svc, _ := newStatusService(role, &grantPermRepo{})

	_, err := svc.ChangeRoleStatus(context.Background(), nil, role.ID, entity.RoleStatusInactive)
	if !sharedErrors.Is(err, sharedErrors.CodeConflict) {
		t.Fatalf("err = %v, want conflict", err)
	}
}

func TestChangeRoleStatusReactivatingProtectedIsAllowed(t *testing.T) {
	role := &entity.Role{ID: uuid.New(), Name: entity.RoleSuperadmin, IsSystem: true, Status: entity.RoleStatusInactive}
	svc, _ := newStatusService(role, &grantPermRepo{})

	got, err := svc.ChangeRoleStatus(context.Background(), nil, role.ID, entity.RoleStatusActive)
	if err != nil {
		t.Fatalf("ChangeRoleStatus: %v", err)
	}
	if got.Status != entity.RoleStatusActive {
		t.Fatalf("status = %q, want active", got.Status)
	}
}

func TestChangeRoleStatusNotFound(t *testing.T) {
	svc, _ := newStatusService(nil, &grantPermRepo{})
	_, err := svc.ChangeRoleStatus(context.Background(), nil, uuid.New(), entity.RoleStatusInactive)
	if !sharedErrors.Is(err, sharedErrors.CodeNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestChangeRoleStatusHidesOtherCompanyRole(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	role := &entity.Role{ID: uuid.New(), CompanyID: &owner, Name: "Vendedor", Status: entity.RoleStatusActive}
	svc, _ := newStatusService(role, &grantPermRepo{})

	_, err := svc.ChangeRoleStatus(context.Background(), &other, role.ID, entity.RoleStatusInactive)
	if !sharedErrors.Is(err, sharedErrors.CodeNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestChangeRoleStatusInvalidStatus(t *testing.T) {
	role := &entity.Role{ID: uuid.New(), Name: "Vendedor", Status: entity.RoleStatusActive}
	svc, _ := newStatusService(role, &grantPermRepo{})

	_, err := svc.ChangeRoleStatus(context.Background(), nil, role.ID, entity.RoleStatus("blocked"))
	if !sharedErrors.Is(err, sharedErrors.CodeValidation) {
		t.Fatalf("err = %v, want validation", err)
	}
}

func TestPermissionCodesByRoleDoesNotReadInactiveRole(t *testing.T) {
	role := &entity.Role{ID: uuid.New(), Name: "Vendedor", Status: entity.RoleStatusInactive}
	perm := &grantPermRepo{}
	svc, _ := newStatusService(role, perm)

	codes, err := svc.PermissionCodesByRole(context.Background(), role.ID)
	if err != nil {
		t.Fatalf("PermissionCodesByRole: %v", err)
	}
	if len(codes) != 0 {
		t.Fatalf("codes = %v, want none", codes)
	}
	if perm.listed != 0 {
		t.Fatalf("listed = %d, want 0 reads for an inactive role", perm.listed)
	}
}

func TestRegisterRejectsInactiveRole(t *testing.T) {
	companyID := uuid.New()
	inactive := &entity.Role{ID: uuid.New(), Name: entity.RoleEmployee, Status: entity.RoleStatusInactive}
	svc := NewUserService(
		newMemoryUserRepo(),
		&memoryRoleRepo{byName: map[string]*entity.Role{entity.RoleEmployee: inactive}},
		&grantPermRepo{},
		clock.NewMockClock(time.Now()),
		&memoryCompanyRepo{ids: map[uuid.UUID]bool{companyID: true}},
	)

	_, err := svc.Register(context.Background(), RegisterUserInput{
		CompanyID: &companyID,
		RoleName:  entity.RoleEmployee,
		FirstName: "Ana",
		LastName:  "Pérez",
		Email:     "ana@example.com",
	})
	if !sharedErrors.Is(err, sharedErrors.CodeValidation) {
		t.Fatalf("err = %v, want validation", err)
	}
}

func TestEditRejectsInactiveRole(t *testing.T) {
	inactive := &entity.Role{ID: uuid.New(), Name: entity.RoleEmployee, Status: entity.RoleStatusInactive}
	existing := activeUser()
	existing.Email = "ana@example.com"
	existing.Username = "ana@example.com"
	users := newMemoryUserRepo(existing)
	svc := NewUserService(
		users,
		&memoryRoleRepo{byName: map[string]*entity.Role{entity.RoleEmployee: inactive}},
		&grantPermRepo{},
		clock.NewMockClock(time.Now()),
		&memoryCompanyRepo{},
	)

	roleName := entity.RoleEmployee
	_, err := svc.Edit(context.Background(), existing.ID, EditUserInput{RoleName: &roleName})
	if !sharedErrors.Is(err, sharedErrors.CodeValidation) {
		t.Fatalf("err = %v, want validation", err)
	}
}

func TestHasPermissionRequiresActiveRole(t *testing.T) {
	tests := []struct {
		name     string
		user     *entity.AppUser
		role     *entity.Role
		granted  bool
		expected bool
	}{
		{name: "active user and active role", user: activeUser(), role: activeRole(), granted: true, expected: true},
		{name: "inactive role grants nothing", user: activeUser(), role: inactiveRole(), granted: true, expected: false},
		{name: "inactive user grants nothing", user: inactiveUser(), role: activeRole(), granted: true, expected: false},
		{name: "missing role grants nothing", user: activeUser(), role: nil, granted: true, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := newMemoryUserRepo(tt.user)
			role := tt.role
			if role != nil {
				role.ID = tt.user.RoleID
			}
			roles := &memoryRoleRepo{byName: map[string]*entity.Role{}}
			if role != nil {
				roles.byName[role.Name] = role
			}
			perm := &grantPermRepo{granted: tt.granted}
			svc := NewUserService(users, roles, perm, clock.NewMockClock(time.Now()), &memoryCompanyRepo{})

			allowed, err := svc.HasPermission(context.Background(), tt.user.ID, "roles", "list")
			if err != nil {
				t.Fatalf("HasPermission: %v", err)
			}
			if allowed != tt.expected {
				t.Fatalf("allowed = %v, want %v", allowed, tt.expected)
			}
		})
	}
}

func activeUser() *entity.AppUser {
	return &entity.AppUser{ID: uuid.New(), RoleID: uuid.New(), Status: entity.UserStatusActive}
}

func inactiveUser() *entity.AppUser {
	u := activeUser()
	u.Status = entity.UserStatusInactive
	return u
}

func activeRole() *entity.Role {
	return &entity.Role{ID: uuid.New(), Name: "Vendedor", Status: entity.RoleStatusActive}
}

func inactiveRole() *entity.Role {
	return &entity.Role{ID: uuid.New(), Name: "Vendedor", Status: entity.RoleStatusInactive}
}

func activeUserWithRole(roleID uuid.UUID) *entity.AppUser {
	return &entity.AppUser{ID: uuid.New(), RoleID: roleID, Status: entity.UserStatusActive}
}
