package service

import (
	"context"

	"github.com/google/uuid"

	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// Delete aplica baja lógica (soft delete) al cliente.
// Nunca elimina físicamente el registro; el índice parcial de unicidad
// permite reutilizar el documento de un cliente dado de baja.
func (s *CustomerService) Delete(ctx context.Context, tenantID uuid.UUID, id int64) error {
	customer, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if customer == nil {
		return sharedErrors.New(sharedErrors.CodeNotFound, "customer not found", 404)
	}

	customer.SoftDelete(s.clock)

	return s.repo.Update(ctx, customer)
}
