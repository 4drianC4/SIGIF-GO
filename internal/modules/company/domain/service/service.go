package service

import (
	"context"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/company/domain/repository"
)

type CompanyService struct{ repo repository.CompanyRepository }

func NewCompanyService(repo repository.CompanyRepository) *CompanyService {
	return &CompanyService{repo: repo}
}
func (s *CompanyService) List(ctx context.Context) ([]entity.Company, error) { return s.repo.List(ctx) }
