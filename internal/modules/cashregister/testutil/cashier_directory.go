package testutil

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/repository"
)

var _ repository.CashierDirectory = (*StubCashierDirectory)(nil)

// StubCashierDirectory stands in for the user module adapter.
type StubCashierDirectory struct {
	mu    sync.Mutex
	names map[uuid.UUID]string
}

func NewStubCashierDirectory() *StubCashierDirectory {
	return &StubCashierDirectory{names: map[uuid.UUID]string{}}
}

func (d *StubCashierDirectory) Add(userID uuid.UUID, name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.names[userID] = name
}

func (d *StubCashierDirectory) Names(_ context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	names := make(map[uuid.UUID]string, len(userIDs))
	for _, id := range userIDs {
		if name, ok := d.names[id]; ok {
			names[id] = name
		}
	}
	return names, nil
}
