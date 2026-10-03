package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// CreateCustomerParams contiene los datos necesarios para crear un cliente.
type CreateCustomerParams struct {
	TenantID       uuid.UUID
	LegalName      string
	DocumentType   entity.DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
}

// Create valida unicidad del documento y persiste el nuevo cliente.
// Reglas:
//  1. LegalName no puede estar vacío después de recortar espacios.
//  2. Si DocumentNumber viene informado, no puede existir otro cliente activo
//     con el mismo (tenant_id, document_type, document_number).
//  3. Los campos financieros, de puntos y el estado inicial los fija la BD.
func (s *CustomerService) Create(ctx context.Context, params CreateCustomerParams) (*entity.Customer, error) {
	// 1. Validar nombre obligatorio
	if strings.TrimSpace(params.LegalName) == "" {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "legal_name is required", 400)
	}

	// 2. Verificar unicidad del documento si viene informado
	if params.DocumentNumber != nil {
		trimmed := strings.TrimSpace(*params.DocumentNumber)
		if trimmed != "" {
			exists, err := s.repo.ExistsByDocument(ctx, params.TenantID, params.DocumentType, trimmed)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, sharedErrors.New(sharedErrors.CodeConflict,
					"a customer with this document already exists in this tenant", 409)
			}
		}
	}

	customer := entity.NewCustomer(
		s.clock,
		params.TenantID,
		params.LegalName,
		params.DocumentType,
		params.DocumentNumber,
		params.Phone,
		params.Email,
	)

	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}
