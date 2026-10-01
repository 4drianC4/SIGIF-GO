package service

import (
	"context"

	"github.com/google/uuid"
)

// RecordLogin marca la hora del último login del usuario.
func (s *UserService) RecordLogin(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	user.RecordLogin(s.clock)
	return s.repo.Update(ctx, user)
}
