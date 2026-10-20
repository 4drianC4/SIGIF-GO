package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/repository"
)

// List returns the active cash registers of the company with their
// availability (HU-10-01, HU-10-02). A cash register in use carries the
// cashier that holds it.
func (s *CashRegisterService) List(
	ctx context.Context,
	companyID uuid.UUID,
	filter repository.ListFilter,
	offset, limit int,
) ([]RegisterView, int64, error) {
	items, total, err := s.repo.ListRegisters(ctx, companyID, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var userIDs []uuid.UUID
	for _, item := range items {
		if item.Assignment != nil {
			userIDs = append(userIDs, item.Assignment.UserID)
		}
	}
	names := map[uuid.UUID]string{}
	if len(userIDs) > 0 {
		names, err = s.cashiers.Names(ctx, userIDs)
		if err != nil {
			return nil, 0, err
		}
	}

	views := make([]RegisterView, len(items))
	for i, item := range items {
		views[i] = RegisterView{Register: item.Register, Assignment: item.Assignment}
		if item.Assignment != nil {
			views[i].CashierName = names[item.Assignment.UserID]
		}
	}
	return views, total, nil
}
