package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

type SortField string

const (
	SortByLegalName SortField = "legal_name"
	SortByCreatedAt SortField = "created_at"
)

func (f SortField) IsValid() bool {
	switch f {
	case SortByLegalName, SortByCreatedAt:
		return true
	}
	return false
}

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

func (o SortOrder) IsValid() bool {
	return o == SortAsc || o == SortDesc
}

// ListFilter narrows a customer listing. A nil Status means every status.
type ListFilter struct {
	Q         string
	Status    *entity.CustomerStatus
	SortBy    SortField
	SortOrder SortOrder
}

type Reader interface {
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Customer, error)

	List(ctx context.Context, companyID uuid.UUID, filter ListFilter, offset, limit int) ([]*entity.Customer, int64, error)

	ExistsByDocument(ctx context.Context, companyID uuid.UUID, docType entity.DocumentType, docNumber string) (bool, error)
}

type Writer interface {
	Create(ctx context.Context, customer *entity.Customer) error

	// Update modifies an existing, non-deleted customer of its company and
	// never inserts; it returns a not-found error when no row matches.
	Update(ctx context.Context, customer *entity.Customer) error
}

type CustomerRepository interface {
	Reader
	Writer

	// WithinTransaction runs fn in a single transaction; repository calls made
	// with the context passed to fn take part in it.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
