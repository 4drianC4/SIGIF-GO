package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type CategoryListItem struct {
	Category       *entity.Category
	ProductsCount  int64
	DefaultTaxName *string
}

type CategoryRepository interface {
	// WithinTransaction runs fn in a single transaction; repository calls made
	// with the context passed to fn take part in it.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	Create(ctx context.Context, category *entity.Category) error
	Update(ctx context.Context, category *entity.Category) error

	// GetByID returns a category that is not deleted, or nil. Inside a
	// transaction the row is locked against deletion until it ends, so a
	// category cannot be deleted while a product or subcategory is being
	// created in it.
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Category, error)

	// GetByIDForUpdate is GetByID with an exclusive lock: it waits for the
	// transactions that read the category and blocks them until it ends.
	GetByIDForUpdate(ctx context.Context, companyID, id uuid.UUID) (*entity.Category, error)

	// GetListItem returns a category with its products count and default tax
	// name, or nil.
	GetListItem(ctx context.Context, companyID, id uuid.UUID) (*CategoryListItem, error)

	// ExistsByName reports whether a category of the same level has the name,
	// ignoring case and skipping excludeID. Deleted categories free their name.
	ExistsByName(ctx context.Context, companyID uuid.UUID, parentID *uuid.UUID, name string, excludeID uuid.UUID) (bool, error)

	// CountSubcategories counts the subcategories that are not deleted, in any
	// status. CountProducts counts the products that are not deleted, in any
	// status.
	CountSubcategories(ctx context.Context, companyID, id uuid.UUID) (int64, error)
	CountProducts(ctx context.Context, companyID, id uuid.UUID) (int64, error)

	// SoftDelete marks the category as deleted. It returns false when there
	// was nothing to delete.
	SoftDelete(ctx context.Context, companyID, id uuid.UUID) (bool, error)

	List(ctx context.Context, companyID uuid.UUID) ([]CategoryListItem, error)
}
