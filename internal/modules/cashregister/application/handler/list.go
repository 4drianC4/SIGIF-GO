package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/query"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
)

func (h *CashRegisterQueryHandler) HandleList(ctx context.Context, q query.ListCashRegisters) ([]service.RegisterView, int64, error) {
	return h.service.List(ctx, q.CompanyID, repository.ListFilter{Availability: q.Availability}, q.Offset, q.Limit)
}
