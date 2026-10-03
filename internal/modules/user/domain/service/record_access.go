package service

import (
	"context"

	"github.com/google/uuid"
)

// RecordAccess updates the user's last access timestamp.
func (s *UserService) RecordAccess(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	user.RecordAccess(s.clock)
	return s.repo.Update(ctx, user)
}
