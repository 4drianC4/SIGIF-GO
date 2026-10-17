package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/cashsession/application/query"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
)

func (h *CashSessionQueryHandler) HandleCurrent(ctx context.Context, q query.GetCurrentSession) (*service.SessionView, error) {
	return h.service.Current(ctx, q.CompanyID, q.UserID)
}
