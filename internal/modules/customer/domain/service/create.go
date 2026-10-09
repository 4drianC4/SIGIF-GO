package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type CreateCustomerParams struct {
	CompanyID      uuid.UUID
	LegalName      string
	DocumentType   entity.DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
}

func (s *CustomerService) Create(ctx context.Context, params CreateCustomerParams) (*entity.Customer, error) {
	if strings.TrimSpace(params.LegalName) == "" {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "legal_name is required", 400)
	}

	if params.DocumentNumber != nil {
		trimmed := strings.TrimSpace(*params.DocumentNumber)
		if trimmed != "" {
			exists, err := s.repo.ExistsByDocument(ctx, params.CompanyID, params.DocumentType, trimmed)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, sharedErrors.New(sharedErrors.CodeConflict,
					"a customer with this document already exists in this company", 409)
			}
		}
	}

	customer := entity.NewCustomer(
		s.clock,
		params.CompanyID,
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
