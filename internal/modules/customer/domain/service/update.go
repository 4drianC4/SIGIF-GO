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
	CompanyID      uuid.UUID
	LegalName      string
	DocumentType   entity.DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
	Address        *string
}

// PatchCustomerParams carries a partial update: a nil field keeps the current
// value. For optional fields an empty string clears the stored value.
type PatchCustomerParams struct {
	ID             uuid.UUID
	CompanyID      uuid.UUID
	LegalName      *string
	DocumentType   *entity.DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
	Address        *string
}

// Update replaces every editable field (PUT).
func (s *CustomerService) Update(ctx context.Context, params UpdateCustomerParams) (*entity.Customer, error) {
	if strings.TrimSpace(params.LegalName) == "" {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "legal_name is required", 400)
	}

	return s.update(ctx, params.CompanyID, params.ID, func(*entity.Customer) UpdateCustomerParams {
		return params
	})
}

// Patch changes only the fields present in params (PATCH).
func (s *CustomerService) Patch(ctx context.Context, params PatchCustomerParams) (*entity.Customer, error) {
	return s.update(ctx, params.CompanyID, params.ID, func(current *entity.Customer) UpdateCustomerParams {
		merged := UpdateCustomerParams{
			ID:             current.ID,
			CompanyID:      current.CompanyID,
			LegalName:      current.LegalName,
			DocumentType:   current.DocumentType,
			DocumentNumber: current.DocumentNumber,
			Phone:          current.Phone,
			Email:          current.Email,
			Address:        current.Address,
		}
		if params.LegalName != nil {
			merged.LegalName = *params.LegalName
		}
		if params.DocumentType != nil {
			merged.DocumentType = *params.DocumentType
		}
		if params.DocumentNumber != nil {
			merged.DocumentNumber = params.DocumentNumber
		}
		if params.Phone != nil {
			merged.Phone = params.Phone
		}
		if params.Email != nil {
			merged.Email = params.Email
		}
		if params.Address != nil {
			merged.Address = params.Address
		}
		return merged
	})
}

// update loads the customer of the company, builds the new values from it and
// saves them in one transaction. The document uniqueness check skips the
// customer itself; the unique index covers concurrent requests.
func (s *CustomerService) update(
	ctx context.Context,
	companyID, id uuid.UUID,
	build func(current *entity.Customer) UpdateCustomerParams,
) (*entity.Customer, error) {
	var customer *entity.Customer

	err := s.repo.WithinTransaction(ctx, func(ctx context.Context) error {
		current, err := s.repo.GetByID(ctx, companyID, id)
		if err != nil {
			return err
		}
		if current == nil {
			return ErrCustomerNotFound
		}

		params := build(current)
		if strings.TrimSpace(params.LegalName) == "" {
			return sharedErrors.New(sharedErrors.CodeValidation, "legal_name is required", 400)
		}

		newDocNumber := ""
		if params.DocumentNumber != nil {
			newDocNumber = strings.TrimSpace(*params.DocumentNumber)
		}
		currentDocNumber := ""
		if current.DocumentNumber != nil {
			currentDocNumber = *current.DocumentNumber
		}

		documentChanged := newDocNumber != currentDocNumber ||
			params.DocumentType != current.DocumentType

		if documentChanged && newDocNumber != "" {
			exists, err := s.repo.ExistsByDocument(ctx, companyID, params.DocumentType, newDocNumber)
			if err != nil {
				return err
			}
			if exists {
				return ErrDocumentTaken
			}
		}

		current.Update(
			s.clock,
			params.LegalName,
			params.DocumentType,
			params.DocumentNumber,
			params.Phone,
			params.Email,
			params.Address,
		)

		if err := s.repo.Update(ctx, current); err != nil {
			return err
		}
		customer = current
		return nil
	})
	if err != nil {
		return nil, err
	}

	return customer, nil
}
