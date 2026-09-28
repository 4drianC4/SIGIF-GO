package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/errors"
)

type TenantService struct {
	repo repository.TenantRepository
}

func NewTenantService(repo repository.TenantRepository) *TenantService {
	return &TenantService{repo: repo}
}

func (s *TenantService) Create(ctx context.Context, name, slug string, businessType entity.BusinessType) (*entity.Tenant, error) {
	exists, err := s.repo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New(errors.CodeConflict, "tenant with this slug already exists", 409)
	}

	tenant := entity.NewTenant(nil, name, slug, businessType)
	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *TenantService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, errors.New(errors.CodeNotFound, "tenant not found", 404)
	}
	return tenant, nil
}

func (s *TenantService) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	tenant, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, errors.New(errors.CodeNotFound, "tenant not found", 404)
	}
	return tenant, nil
}

func (s *TenantService) List(ctx context.Context, offset, limit int) ([]*entity.Tenant, int64, error) {
	return s.repo.List(ctx, offset, limit)
}

func (s *TenantService) Update(ctx context.Context, id uuid.UUID, name string, settings entity.TenantSettings) (*entity.Tenant, error) {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, errors.New(errors.CodeNotFound, "tenant not found", 404)
	}

	tenant.Update(name, settings, nil)
	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *TenantService) Delete(ctx context.Context, id uuid.UUID) error {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tenant == nil {
		return errors.New(errors.CodeNotFound, "tenant not found", 404)
	}
	return s.repo.Delete(ctx, id)
}

func (s *TenantService) Activate(ctx context.Context, id uuid.UUID) error {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tenant == nil {
		return errors.New(errors.CodeNotFound, "tenant not found", 404)
	}
	tenant.Activate(nil)
	return s.repo.Update(ctx, tenant)
}

func (s *TenantService) Deactivate(ctx context.Context, id uuid.UUID) error {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tenant == nil {
		return errors.New(errors.CodeNotFound, "tenant not found", 404)
	}
	tenant.Deactivate(nil)
	return s.repo.Update(ctx, tenant)
}