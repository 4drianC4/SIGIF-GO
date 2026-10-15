package service

import (
	"context"

	"github.com/google/uuid"
)

// Current returns the cash register the cashier holds during the shift
// (HU-10-01).
func (s *CashRegisterService) Current(ctx context.Context, companyID, userID uuid.UUID) (*AssignmentView, error) {
	assignment, err := s.repo.ActiveAssignmentByUser(ctx, companyID, userID)
	if err != nil {
		return nil, err
	}
	if assignment == nil {
		return nil, ErrNoActiveAssignment
	}
	return s.assignmentView(ctx, assignment)
}
