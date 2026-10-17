package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

type grantAllPermRepo struct{}

func (grantAllPermRepo) HasPermission(_ context.Context, _ uuid.UUID, _, _ string) (bool, error) {
	return true, nil
}

func (grantAllPermRepo) List(context.Context, string, int, int) ([]*entity.Permission, int64, error) {
	return nil, 0, nil
}

func (grantAllPermRepo) GetByModuleOperation(context.Context, string, string) (*entity.Permission, error) {
	return nil, nil
}

func (grantAllPermRepo) GetByID(context.Context, uuid.UUID) (*entity.Permission, error) {
	return nil, nil
}

func (grantAllPermRepo) ListAll(context.Context, string) ([]*entity.Permission, error) {
	return nil, nil
}

func (grantAllPermRepo) ListByRole(context.Context, uuid.UUID) ([]*entity.Permission, error) {
	return nil, nil
}

func (grantAllPermRepo) Create(context.Context, *entity.Permission) error { return nil }

func (grantAllPermRepo) Update(context.Context, *entity.Permission) error { return nil }

func (grantAllPermRepo) Delete(context.Context, uuid.UUID) error { return nil }

func (grantAllPermRepo) CountRolesByPermission(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}

func seedStoredUser(t *testing.T, users *memoryUserRepo, email string) *entity.AppUser {
	t.Helper()
	user := seedTestUser(t, email)
	users.byID[user.ID] = user
	users.byEmail[user.Email] = user
	return user
}

func TestDeactivateMarksUserInactiveAndSoftDeleted(t *testing.T) {
	svc, users, _ := setupEditService()
	user := seedStoredUser(t, users, "juan@sigif.com")
	hash := user.PasswordHash

	if err := svc.Deactivate(context.Background(), user.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	stored := users.byID[user.ID]
	if stored == nil {
		t.Fatal("user must not be physically removed")
	}
	if stored.Status != entity.UserStatusInactive {
		t.Fatalf("status = %q, want inactive", stored.Status)
	}
	if stored.DeletedAt == nil {
		t.Fatal("deleted_at must be set")
	}
	if stored.IsActive() {
		t.Fatal("deactivated user must not be active")
	}
	if !stored.UpdatedAt.Equal(*stored.DeletedAt) {
		t.Fatalf("updated_at %v must match deleted_at %v", stored.UpdatedAt, *stored.DeletedAt)
	}
	if stored.Email != "juan@sigif.com" || stored.FirstName != "Ana" || stored.PasswordHash != hash {
		t.Fatalf("deactivation must not change other data: %+v", stored)
	}
}

func TestDeactivateAlreadyInactiveUserIsConflict(t *testing.T) {
	svc, users, _ := setupEditService()
	user := seedStoredUser(t, users, "juan@sigif.com")
	ctx := context.Background()

	if err := svc.Deactivate(ctx, user.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	firstDeletedAt := *users.byID[user.ID].DeletedAt

	err := svc.Deactivate(ctx, user.ID)
	if !sharedErrors.Is(err, sharedErrors.CodeConflict) {
		t.Fatalf("want CONFLICT, got %v", err)
	}
	if !users.byID[user.ID].DeletedAt.Equal(firstDeletedAt) {
		t.Fatal("a rejected deactivation must not change deleted_at")
	}
}

func TestDeactivateUnknownUserIsNotFound(t *testing.T) {
	svc, _, _ := setupEditService()

	err := svc.Deactivate(context.Background(), uuid.New())
	if !sharedErrors.Is(err, sharedErrors.CodeNotFound) {
		t.Fatalf("want NOT_FOUND, got %v", err)
	}
}

func TestDeactivateOwnAccountIsConflict(t *testing.T) {
	svc, users, _ := setupEditService()
	user := seedStoredUser(t, users, "admin@sigif.com")
	ctx := middleware.WithUserID(context.Background(), user.ID)

	err := svc.Deactivate(ctx, user.ID)
	if !sharedErrors.Is(err, sharedErrors.CodeConflict) {
		t.Fatalf("want CONFLICT, got %v", err)
	}
	if stored := users.byID[user.ID]; stored.Status != entity.UserStatusActive || stored.DeletedAt != nil {
		t.Fatalf("own account must stay active: %+v", stored)
	}
}

func TestDeactivateByAnotherUserIsAllowed(t *testing.T) {
	svc, users, _ := setupEditService()
	user := seedStoredUser(t, users, "juan@sigif.com")
	ctx := middleware.WithUserID(context.Background(), uuid.New())

	if err := svc.Deactivate(ctx, user.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
}

func TestActivateRestoresDeactivatedUser(t *testing.T) {
	svc, users, _ := setupEditService()
	user := seedStoredUser(t, users, "juan@sigif.com")
	ctx := context.Background()

	if err := svc.Deactivate(ctx, user.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if err := svc.Activate(ctx, user.ID); err != nil {
		t.Fatalf("activate: %v", err)
	}

	stored := users.byID[user.ID]
	if stored.Status != entity.UserStatusActive || stored.DeletedAt != nil || !stored.IsActive() {
		t.Fatalf("user must be active again: %+v", stored)
	}
	if err := svc.Deactivate(ctx, user.ID); err != nil {
		t.Fatalf("a reactivated user can be deactivated again: %v", err)
	}
}

func TestDeactivatedUserLosesPermissions(t *testing.T) {
	users := newMemoryUserRepo()
	user := seedStoredUser(t, users, "juan@sigif.com")
	roles := &memoryRoleRepo{byName: map[string]*entity.Role{
		entity.RoleEmployee: {ID: user.RoleID, Name: entity.RoleEmployee, Status: entity.RoleStatusActive},
	}}
	svc := NewUserService(users, roles, grantAllPermRepo{}, clock.NewMockClock(time.Now()), &memoryCompanyRepo{})
	ctx := context.Background()

	allowed, err := svc.HasPermission(ctx, user.ID, "users", "read")
	if err != nil || !allowed {
		t.Fatalf("active user must be allowed: %v %v", allowed, err)
	}
	if err := svc.Deactivate(ctx, user.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	allowed, err = svc.HasPermission(ctx, user.ID, "users", "read")
	if err != nil || allowed {
		t.Fatalf("deactivated user must lose permissions: %v %v", allowed, err)
	}
}
