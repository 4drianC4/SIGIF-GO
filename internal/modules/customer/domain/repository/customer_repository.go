package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

type ListFilter struct {
	Q      string
	Status *entity.CustomerStatus
}

type Reader interface {
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Customer, error)

	List(ctx context.Context, companyID uuid.UUID, filter ListFilter, offset, limit int) ([]*entity.Customer, int64, error)

	ExistsByDocument(ctx context.Context, companyID uuid.UUID, docType entity.DocumentType, docNumber string) (bool, error)
}

type Writer interface {
	Create(ctx context.Context, customer *entity.Customer) error

	Update(ctx context.Context, customer *entity.Customer) error
}

type CustomerRepository interface {
	Reader
	Writer
}
