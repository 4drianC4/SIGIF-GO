package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
)

// List retorna una página de clientes del tenant aplicando los filtros indicados.
// El total corresponde al conjunto completo tras aplicar filtros (antes del offset/limit).
func (s *CustomerService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	filter repository.ListFilter,
	offset, limit int,
) ([]*entity.Customer, int64, error) {
	return s.repo.List(ctx, tenantID, filter, offset, limit)
}
