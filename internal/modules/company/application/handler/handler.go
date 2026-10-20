package handler

import (
	"context"
	"github.com/sigif/sigif-go/internal/modules/company/application/dto"
	"github.com/sigif/sigif-go/internal/modules/company/domain/service"
)

type CompanyQueryHandler struct{ service *service.CompanyService }

func NewCompanyQueryHandler(s *service.CompanyService) *CompanyQueryHandler {
	return &CompanyQueryHandler{service: s}
}
func (h *CompanyQueryHandler) HandleList(ctx context.Context) ([]dto.CompanyOption, error) {
	companies, err := h.service.List(ctx)
	if err != nil {
		return nil, err
	}
	options := make([]dto.CompanyOption, len(companies))
	for i, c := range companies {
		options[i] = dto.CompanyOption{ID: c.ID.String(), LegalName: c.LegalName, TradeName: c.TradeName}
	}
	return options, nil
}
