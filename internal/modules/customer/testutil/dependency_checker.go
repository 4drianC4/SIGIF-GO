package testutil

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
)

var _ repository.DependencyChecker = (*StubDependencyChecker)(nil)

// StubDependencyChecker reports dependencies for the customers marked with
// Block, standing in for the sales, credit and loyalty modules.
type StubDependencyChecker struct {
	mu      sync.Mutex
	blocked map[uuid.UUID]bool
}

func NewStubDependencyChecker() *StubDependencyChecker {
	return &StubDependencyChecker{blocked: map[uuid.UUID]bool{}}
}

func (d *StubDependencyChecker) Block(customerID uuid.UUID) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.blocked[customerID] = true
}

func (d *StubDependencyChecker) HasDependencies(_ context.Context, _, customerID uuid.UUID) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.blocked[customerID], nil
}
