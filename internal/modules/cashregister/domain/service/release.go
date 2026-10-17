package service

import (
	"context"

	"github.com/google/uuid"
)

type ReleaseParams struct {
	CompanyID    uuid.UUID
	ActorID      uuid.UUID
	AssignmentID uuid.UUID
	// CanReleaseAny lets a supervisor release an assignment of another cashier.
	CanReleaseAny bool
}

// Release ends an assignment so the cash register can be assigned again
// (HU-10-02, criterion 5). Only the assigned cashier or a supervisor may
// release it. Releasing an assignment that is already released is idempotent:
// it is returned without writing to the database.
func (s *CashRegisterService) Release(ctx context.Context, params ReleaseParams) (*AssignmentView, error) {
	assignment, err := s.repo.GetAssignment(ctx, params.CompanyID, params.AssignmentID)
	if err != nil {
		return nil, err
	}
	if assignment == nil {
		return nil, ErrAssignmentNotFound
	}
	if !assignment.BelongsTo(params.ActorID) && !params.CanReleaseAny {
		return nil, ErrReleaseForbidden
	}
	if !assignment.IsActive() {
		return s.assignmentView(ctx, assignment)
	}

	assignment.Release(s.clock, params.ActorID)
	released, err := s.repo.ReleaseAssignment(ctx, assignment)
	if err != nil {
		return nil, err
	}
	if !released {
		// Another request released it first: return the stored release.
		assignment, err = s.repo.GetAssignment(ctx, params.CompanyID, params.AssignmentID)
		if err != nil {
			return nil, err
		}
		if assignment == nil {
			return nil, ErrAssignmentNotFound
		}
	}

	return s.assignmentView(ctx, assignment)
}
