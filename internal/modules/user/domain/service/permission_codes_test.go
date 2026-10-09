package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

// codesPermRepo reuses the no-op repository and only stubs the by-role listing
// needed by PermissionCodesByRole.
type codesPermRepo struct {
	noopPermRepo
	permissions []*entity.Permission
}

func (r codesPermRepo) ListByRole(_ context.Context, _ uuid.UUID) ([]*entity.Permission, error) {
	return r.permissions, nil
}

func TestPermissionCodesByRole(t *testing.T) {
	repo := codesPermRepo{permissions: []*entity.Permission{
		{ID: uuid.New(), Module: "permissions", Operation: "read"},
		{ID: uuid.New(), Module: "users", Operation: "list"},
	}}
	svc := NewUserService(
		newMemoryUserRepo(),
		&memoryRoleRepo{byName: map[string]*entity.Role{}},
		repo,
		clock.NewMockClock(time.Now()),
		&memoryCompanyRepo{},
	)

	codes, err := svc.PermissionCodesByRole(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("PermissionCodesByRole() error = %v", err)
	}
	if len(codes) != 2 || codes[0] != "permissions.read" || codes[1] != "users.list" {
		t.Fatalf("codes = %v, want [permissions.read users.list]", codes)
	}
}
