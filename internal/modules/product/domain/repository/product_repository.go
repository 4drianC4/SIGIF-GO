package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

// ProductRepository es el puerto de persistencia de productos.
type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	ExistsBySKU(ctx context.Context, tenantID uuid.UUID, sku string) (bool, error)
	ExistsByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (bool, error)
}
