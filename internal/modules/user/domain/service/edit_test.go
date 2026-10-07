package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/security"
)

type memoryUserRepo struct {
	byID    map[uuid.UUID]*entity.AppUser
	byEmail map[string]*entity.AppUser
}

func newMemoryUserRepo(users ...*entity.AppUser) *memoryUserRepo {
	r := &memoryUserRepo{byID: map[uuid.UUID]*entity.AppUser{}, byEmail: map[string]*entity.AppUser{}}
	for _, u := range users {
		r.byID[u.ID] = u
		r.byEmail[u.Email] = u
	}
	return r
}

func (r *memoryUserRepo) Create(_ context.Context, u *entity.AppUser) error {
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u
	return nil
}

func (r *memoryUserRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.AppUser, error) {
	return r.byID[id], nil
}

func (r *memoryUserRepo) GetByEmail(_ context.Context, email string) (*entity.AppUser, error) {
	return r.byEmail[email], nil
}

func (r *memoryUserRepo) List(_ context.Context, _, _ int) ([]*entity.AppUser, int64, error) {
	out := make([]*entity.AppUser, 0, len(r.byID))
	for _, u := range r.byID {
		out = append(out, u)
	}
	return out, int64(len(out)), nil
}

func (r *memoryUserRepo) Update(_ context.Context, u *entity.AppUser) error {
	// keep the byEmail index in sync for uniqueness checks
	for e, existing := range r.byEmail {
		if existing.ID == u.ID {
			delete(r.byEmail, e)
		}
	}
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u
	return nil
}

func (r *memoryUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.byID, id)
	return nil
}

func (r *memoryUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	_, ok := r.byEmail[email]
	return ok, nil
}

func (r *memoryUserRepo) ExistsByID(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := r.byID[id]
	return ok, nil
}

type memoryRoleRepo struct {
	byName map[string]*entity.Role
}

func (r *memoryRoleRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Role, error) {
	for _, role := range r.byName {
		if role.ID == id {
			return role, nil
		}
	}
	return nil, nil
}

func (r *memoryRoleRepo) GetByName(_ context.Context, name string) (*entity.Role, error) {
	return r.byName[name], nil
}

type noopPermRepo struct{}

func (noopPermRepo) HasPermission(_ context.Context, _ uuid.UUID, _, _ string) (bool, error) {
	return false, nil
}

func (noopPermRepo) List(_ context.Context, _ string, _, _ int) ([]*entity.Permission, int64, error) {
	return nil, 0, nil
}

func (noopPermRepo) GetByModuleOperation(_ context.Context, _, _ string) (*entity.Permission, error) {
	return nil, nil
}

func (noopPermRepo) Create(_ context.Context, _ *entity.Permission) error {
	return nil
}

func seedTestUser(t *testing.T, email string) *entity.AppUser {
	t.Helper()
	hash, err := security.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return &entity.AppUser{
		ID:                uuid.New(),
		RoleID:            uuid.New(),
		FirstName:         "Ana",
		LastName:          "Pérez",
		Username:          email,
		Email:             email,
		PasswordHash:      hash,
		PasswordAlgorithm: security.Algorithm,
		Status:            entity.UserStatusActive,
	}
}

func setupEditService() (*UserService, *memoryUserRepo, *memoryRoleRepo) {
	users := newMemoryUserRepo()
	roles := &memoryRoleRepo{byName: map[string]*entity.Role{
		"superadmin": {ID: uuid.New(), Name: "superadmin"},
		"soporte":    {ID: uuid.New(), Name: "soporte"},
	}}
	svc := NewUserService(users, roles, noopPermRepo{}, clock.NewMockClock(time.Now()))
	return svc, users, roles
}

func strPtr(s string) *string { return &s }

func TestEditAppliesOnlyProvidedFields(t *testing.T) {
	svc, users, roles := setupEditService()
	existing := seedTestUser(t, "ana@sigif.com")
	existing.Area = "Old area"
	users.byID[existing.ID] = existing
	users.byEmail[existing.Email] = existing

	soporteID := roles.byName["soporte"].ID
	got, err := svc.Edit(context.Background(), existing.ID, EditUserInput{
		FirstName: strPtr("Ana María"),
		RoleName:  strPtr("soporte"),
	})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got.FirstName != "Ana María" {
		t.Errorf("first_name = %q, want changed", got.FirstName)
	}
	if got.LastName != "Pérez" {
		t.Errorf("last_name = %q, want unchanged", got.LastName)
	}
	if got.RoleID != soporteID {
		t.Errorf("role_id not updated to soporte")
	}
	if got.Area != "Old area" {
		t.Errorf("area = %q, want unchanged", got.Area)
	}
}

func TestEditEmailConflict(t *testing.T) {
	svc, users, _ := setupEditService()
	a := seedTestUser(t, "a@sigif.com")
	b := seedTestUser(t, "b@sigif.com")
	users.byID[a.ID] = a
	users.byEmail[a.Email] = a
	users.byID[b.ID] = b
	users.byEmail[b.Email] = b

	_, err := svc.Edit(context.Background(), a.ID, EditUserInput{Email: strPtr("b@sigif.com")})
	if err == nil || !sharedErrors.Is(err, sharedErrors.CodeConflict) {
		t.Fatalf("expected CONFLICT, got %v", err)
	}
}

func TestEditEmailUpdatesUsername(t *testing.T) {
	svc, users, _ := setupEditService()
	a := seedTestUser(t, "a@sigif.com")
	users.byID[a.ID] = a
	users.byEmail[a.Email] = a

	got, err := svc.Edit(context.Background(), a.ID, EditUserInput{Email: strPtr("nuevo@sigif.com")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got.Email != "nuevo@sigif.com" || got.Username != "nuevo@sigif.com" {
		t.Errorf("email/username not synced: %q / %q", got.Email, got.Username)
	}
}

func TestEditInvalidRole(t *testing.T) {
	svc, users, _ := setupEditService()
	a := seedTestUser(t, "a@sigif.com")
	users.byID[a.ID] = a
	users.byEmail[a.Email] = a

	_, err := svc.Edit(context.Background(), a.ID, EditUserInput{RoleName: strPtr("ghost")})
	if err == nil || !sharedErrors.Is(err, sharedErrors.CodeValidation) {
		t.Fatalf("expected VALIDATION_ERROR, got %v", err)
	}
}

func TestEditNotFound(t *testing.T) {
	svc, _, _ := setupEditService()
	_, err := svc.Edit(context.Background(), uuid.New(), EditUserInput{FirstName: strPtr("X")})
	if err == nil || !sharedErrors.Is(err, sharedErrors.CodeNotFound) {
		t.Fatalf("expected NOT_FOUND, got %v", err)
	}
}

func TestEditPasswordRehashes(t *testing.T) {
	svc, users, _ := setupEditService()
	a := seedTestUser(t, "a@sigif.com")
	users.byID[a.ID] = a
	users.byEmail[a.Email] = a

	got, err := svc.Edit(context.Background(), a.ID, EditUserInput{Password: strPtr("newpassword456")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if err := security.VerifyPassword("newpassword456", got.PasswordHash); err != nil {
		t.Errorf("password was not rehashed: %v", err)
	}
}
