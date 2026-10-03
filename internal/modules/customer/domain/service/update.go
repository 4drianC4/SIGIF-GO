package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type UpdateCustomerParams struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	LegalName      string
	DocumentType   entity.DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
	Address        *string
}

func (s *CustomerService) Update(ctx context.Context, params UpdateCustomerParams) (*entity.Customer, error) {
	if strings.TrimSpace(params.LegalName) == "" {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "legal_name is required", 400)
	}

	customer, err := s.repo.GetByID(ctx, params.TenantID, params.ID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, sharedErrors.New(sharedErrors.CodeNotFound, "customer not found", 404)
	}

	newDocNumber := ""
	if params.DocumentNumber != nil {
		newDocNumber = strings.TrimSpace(*params.DocumentNumber)
	}
	currentDocNumber := ""
	if customer.DocumentNumber != nil {
		currentDocNumber = *customer.DocumentNumber
	}

	documentChanged := newDocNumber != currentDocNumber ||
		params.DocumentType != customer.DocumentType

	if documentChanged && newDocNumber != "" {
		exists, err := s.repo.ExistsByDocument(ctx, params.TenantID, params.DocumentType, newDocNumber)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, sharedErrors.New(sharedErrors.CodeConflict,
				"a customer with this document already exists in this tenant", 409)
		}
	}

	customer.Update(
		s.clock,
		params.LegalName,
		params.DocumentType,
		params.DocumentNumber,
		params.Phone,
		params.Email,
		params.Address,
	)

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}
