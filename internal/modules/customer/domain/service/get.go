package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// GetByID retorna el cliente con ese ID dentro del tenant autenticado.
// Un ID válido de otro tenant se trata como no encontrado para no revelar su existencia.
func (s *CustomerService) GetByID(ctx context.Context, tenantID uuid.UUID, id int64) (*entity.Customer, error) {
	customer, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, sharedErrors.New(sharedErrors.CodeNotFound, "customer not found", 404)
	}
	return customer, nil
}
