package gorm

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
)

// NoDependencyChecker finds no dependencies: no table references customer
// yet. The sales, credit and loyalty modules will replace it with a checker
// that looks at their own records.
type NoDependencyChecker struct{}

func NewNoDependencyChecker() repository.DependencyChecker {
	return NoDependencyChecker{}
}

func (NoDependencyChecker) HasDependencies(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}
