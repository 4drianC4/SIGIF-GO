package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/customer/application/query"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
)

func (h *CustomerQueryHandler) HandleGet(ctx context.Context, q query.GetCustomer) (*entity.Customer, error) {
	return h.service.GetByID(ctx, q.TenantID, q.ID)
}

func (h *CustomerQueryHandler) HandleList(ctx context.Context, q query.ListCustomers) ([]*entity.Customer, int64, error) {
	filter := repository.ListFilter{
		Q:      q.Q,
		Status: q.Status,
	}
	return h.service.List(ctx, q.TenantID, filter, q.Offset, q.Limit)
}
