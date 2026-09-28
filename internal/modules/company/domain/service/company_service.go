package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/company/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/errors"
)

type CompanyService struct {
	repo repository.CompanyRepository
}

func NewCompanyService(repo repository.CompanyRepository) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) Create(ctx context.Context, tenantID uuid.UUID, name, legalName, taxID string) (*entity.Company, error) {
	exists, err := s.repo.ExistsByTaxID(ctx, taxID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New(errors.CodeConflict, "company with this tax ID already exists", 409)
	}

	company := entity.NewCompany(nil, tenantID, name, legalName, taxID)
	if err := s.repo.Create(ctx, company); err != nil {
		return nil, err
	}
	return company, nil
}

func (s *CompanyService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Company, error) {
	company, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, errors.New(errors.CodeNotFound, "company not found", 404)
	}
	return company, nil
}

func (s *CompanyService) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Company, error) {
	return s.repo.GetByTenantID(ctx, tenantID)
}

func (s *CompanyService) List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*entity.Company, int64, error) {
	return s.repo.List(ctx, tenantID, offset, limit)
}

func (s *CompanyService) Update(ctx context.Context, id uuid.UUID, name, legalName, taxID, email, phone, address, city, state, country, postalCode string, settings entity.CompanySettings) (*entity.Company, error) {
	company, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, errors.New(errors.CodeNotFound, "company not found", 404)
	}

	company.Update(name, legalName, taxID, email, phone, address, city, state, country, postalCode, settings, nil)
	if err := s.repo.Update(ctx, company); err != nil {
		return nil, err
	}
	return company, nil
}

func (s *CompanyService) Delete(ctx context.Context, id uuid.UUID) error {
	company, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if company == nil {
		return errors.New(errors.CodeNotFound, "company not found", 404)
	}
	return s.repo.Delete(ctx, id)
}

func (s *CompanyService) Activate(ctx context.Context, id uuid.UUID) error {
	company, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if company == nil {
		return errors.New(errors.CodeNotFound, "company not found", 404)
	}
	company.Activate(nil)
	return s.repo.Update(ctx, company)
}

func (s *CompanyService) Deactivate(ctx context.Context, id uuid.UUID) error {
	company, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if company == nil {
		return errors.New(errors.CodeNotFound, "company not found", 404)
	}
	company.Deactivate(nil)
	return s.repo.Update(ctx, company)
}