package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/query"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
)

func (h *CashRegisterQueryHandler) HandleCurrent(ctx context.Context, q query.GetCurrentAssignment) (*service.AssignmentView, error) {
	return h.service.Current(ctx, q.CompanyID, q.UserID)
}
