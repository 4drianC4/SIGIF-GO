package repository

import (
	"context"

	"github.com/google/uuid"
)

// CashierDirectory resolves the display name of cashiers. The user module
// provides it through an adapter.
type CashierDirectory interface {
	Names(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)
}
