package service

import (
	"context"
	"github.com/google/uuid"
	companyEntity "github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/security"
	"testing"
)

type memoryCompanyRepo struct{ ids map[uuid.UUID]bool }

func (r *memoryCompanyRepo) ExistsByID(_ context.Context, id uuid.UUID) (bool, error) {
	return r.ids[id], nil
}
func (r *memoryCompanyRepo) List(context.Context) ([]companyEntity.Company, error) {
	return []companyEntity.Company{}, nil
}
func TestRegisterGeneratesPasswordAndAssignsCompany(t *testing.T) {
	s, users, roles := setupEditService()
	id := uuid.New()
	s.companyRepo = &memoryCompanyRepo{ids: map[uuid.UUID]bool{id: true}}
	input := RegisterUserInput{CompanyID: &id, RoleName: entity.RoleBusinessAdmin, FirstName: "Ana", LastName: "Pérez", Email: "ana@example.com"}
	result, err := s.Register(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	u := result.User
	if u.CompanyID == nil || *u.CompanyID != id || u.RoleName != input.RoleName || u.RoleID != roles.byName[input.RoleName].ID {
		t.Fatalf("incorrect company or role: %+v", u)
	}
	if len(result.GeneratedPassword) < 24 {
		t.Fatal("generated password too short")
	}
	if err := security.VerifyPassword(result.GeneratedPassword, users.byID[u.ID].PasswordHash); err != nil {
		t.Fatal(err)
	}
	input.Email = "second@example.com"
	second, err := s.Register(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if second.GeneratedPassword == result.GeneratedPassword {
		t.Fatal("passwords must differ")
	}
	_, err = s.Register(context.Background(), input)
	if !sharedErrors.Is(err, sharedErrors.CodeConflict) {
		t.Fatalf("expected duplicate conflict: %v", err)
	}
}
func TestRegisterRejectsInvalidCompany(t *testing.T) {
	for _, id := range []*uuid.UUID{nil, ptrUUID(uuid.Nil), ptrUUID(uuid.New())} {
		s, users, _ := setupEditService()
		_, err := s.Register(context.Background(), RegisterUserInput{CompanyID: id, RoleName: entity.RoleBusinessAdmin})
		if !sharedErrors.Is(err, sharedErrors.CodeValidation) || len(users.byID) != 0 {
			t.Fatalf("expected validation without write: %v", err)
		}
	}
}
func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }
func TestEditCompanyAndPreservesOmittedFields(t *testing.T) {
	s, users, _ := setupEditService()
	u := seedTestUser(t, "ana@example.com")
	old, next := uuid.New(), uuid.New()
	u.CompanyID = &old
	users.Create(context.Background(), u)
	s.companyRepo = &memoryCompanyRepo{ids: map[uuid.UUID]bool{next: true}}
	hash := u.PasswordHash
	got, err := s.Edit(context.Background(), u.ID, EditUserInput{CompanyID: &next})
	if err != nil {
		t.Fatal(err)
	}
	if *got.CompanyID != next || got.PasswordHash != hash || got.Email != "ana@example.com" {
		t.Fatal("unexpected patch result")
	}
	got, err = s.Edit(context.Background(), u.ID, EditUserInput{FirstName: strPtr("Changed")})
	if err != nil || *got.CompanyID != next {
		t.Fatalf("omitted company changed: %v", err)
	}
	for _, invalid := range []uuid.UUID{uuid.Nil, uuid.New()} {
		_, err = s.Edit(context.Background(), u.ID, EditUserInput{CompanyID: &invalid, FirstName: strPtr("Rejected")})
		if !sharedErrors.Is(err, sharedErrors.CodeValidation) || u.FirstName != "Changed" || *u.CompanyID != next {
			t.Fatalf("invalid company mutated user: %v", err)
		}
	}
}
