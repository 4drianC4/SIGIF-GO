package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/errors"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, tenantID uuid.UUID, companyID *uuid.UUID, email, password, firstName, lastName string, roles []entity.UserRole) (*entity.User, error) {
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New(errors.CodeConflict, "user with this email already exists", 409)
	}

	user, err := entity.NewUser(nil, &tenantID, companyID, email, password, firstName, lastName, roles)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(errors.CodeNotFound, "user not found", 404)
	}
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(errors.CodeNotFound, "user not found", 404)
	}
	return user, nil
}

func (s *UserService) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error) {
	return s.repo.GetByTenantID(ctx, tenantID)
}

func (s *UserService) GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*entity.User, error) {
	return s.repo.GetByCompanyID(ctx, companyID)
}

func (s *UserService) List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.User, int64, error) {
	return s.repo.List(ctx, tenantID, offset, limit)
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, firstName, lastName, phone, avatarURL string, roles []entity.UserRole, settings entity.UserSettings) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(errors.CodeNotFound, "user not found", 404)
	}

	user.Update(firstName, lastName, phone, avatarURL, roles, settings, nil)
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) ChangePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(errors.CodeNotFound, "user not found", 404)
	}

	if err := user.CheckPassword(currentPassword); err != nil {
		return errors.New(errors.CodeUnauthorized, "current password is incorrect", 401)
	}

	if err := user.ChangePassword(newPassword); err != nil {
		return err
	}

	return s.repo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(errors.CodeNotFound, "user not found", 404)
	}
	return s.repo.Delete(ctx, id)
}

func (s *UserService) Activate(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(errors.CodeNotFound, "user not found", 404)
	}
	user.Activate(nil)
	return s.repo.Update(ctx, user)
}

func (s *UserService) Deactivate(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(errors.CodeNotFound, "user not found", 404)
	}
	user.Deactivate(nil)
	return s.repo.Update(ctx, user)
}

func (s *UserService) Suspend(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(errors.CodeNotFound, "user not found", 404)
	}
	user.Suspend(nil)
	return s.repo.Update(ctx, user)
}

func (s *UserService) RecordLogin(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(errors.CodeNotFound, "user not found", 404)
	}
	user.RecordLogin(nil)
	return s.repo.Update(ctx, user)
}