package repository

import (
	"context"

	"github.com/google/uuid"
)

// DependencyChecker reports whether a customer is referenced by records that
// must be kept (sales, credits, loyalty movements). Modules that own those
// records provide the implementation.
type DependencyChecker interface {
	HasDependencies(ctx context.Context, companyID, customerID uuid.UUID) (bool, error)
}
