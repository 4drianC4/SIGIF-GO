package service

import (
	"context"
	"github.com/google/uuid"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func (s *UserService) validateCompany(ctx context.Context, id *uuid.UUID) error {
	if id == nil || *id == uuid.Nil {
		return sharedErrors.New(sharedErrors.CodeValidation, "company_id is required", 400)
	}
	exists, err := s.companyRepo.ExistsByID(ctx, *id)
	if err != nil {
		return err
	}
	if !exists {
		return sharedErrors.New(sharedErrors.CodeValidation, "company does not exist", 400)
	}
	return nil
}
