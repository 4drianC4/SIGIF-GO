package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Product, error)
	ExistsBySKU(ctx context.Context, companyID uuid.UUID, sku string) (bool, error)
	ExistsBySKUExcluding(ctx context.Context, companyID uuid.UUID, sku string, excludeID uuid.UUID) (bool, error)
	ExistsByBarcode(ctx context.Context, companyID uuid.UUID, barcode string) (bool, error)
	ExistsByBarcodeExcluding(ctx context.Context, companyID uuid.UUID, barcode string, excludeID uuid.UUID) (bool, error)
}
