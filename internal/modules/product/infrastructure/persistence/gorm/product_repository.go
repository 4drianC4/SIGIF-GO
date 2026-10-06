package gorm

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

type ProductGormRepository struct {
	db *sharedDatabase.Database
}

func NewProductGormRepository(db *sharedDatabase.Database) repository.ProductRepository {
	return &ProductGormRepository{db: db}
}

func (r *ProductGormRepository) Create(ctx context.Context, product *entity.Product) error {
	db := r.db.GetDB(ctx)
	if err := db.Create(mapper.ProductToModel(product)).Error; err != nil {
		if isDuplicateKey(db, err) {
			switch {
			case strings.Contains(err.Error(), "idx_products_company_name"):
				return service.ErrProductNameTaken
			case strings.Contains(err.Error(), "idx_products_company_barcode"):
				return service.ErrBarcodeTaken
			default:
				return service.ErrSKUTaken
			}
		}
		return err
	}
	return nil
}

func (r *ProductGormRepository) ExistsByName(ctx context.Context, companyID uuid.UUID, name string) (bool, error) {
	return r.exists(ctx, "company_id = ? AND LOWER(name) = LOWER(?)", companyID, name)
}

func (r *ProductGormRepository) ExistsBySKU(ctx context.Context, companyID uuid.UUID, sku string) (bool, error) {
	return r.exists(ctx, "company_id = ? AND sku = ?", companyID, sku)
}

func (r *ProductGormRepository) ExistsByBarcode(ctx context.Context, companyID uuid.UUID, barcode string) (bool, error) {
	return r.exists(ctx, "company_id = ? AND barcode = ?", companyID, barcode)
}

func (r *ProductGormRepository) exists(ctx context.Context, condition string, args ...any) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.ProductModel{}).Where(condition, args...).Count(&count).Error
	return count > 0, err
}

type productRow struct {
	model.ProductModel `gorm:"embedded"`
	CategoryName       string
}

func (r *ProductGormRepository) List(ctx context.Context, companyID uuid.UUID, filter repository.ProductFilter, offset, limit int) ([]repository.ProductListItem, int64, error) {
	base := func() *gorm.DB {
		db := r.db.GetDB(ctx).Table("products AS p").
			Joins("JOIN categories c ON c.id = p.category_id").
			Where("p.company_id = ? AND p.deleted_at IS NULL", companyID)
		if search := strings.TrimSpace(filter.Search); search != "" {
			pattern := "%" + likeEscaper.Replace(strings.ToLower(search)) + "%"
			db = db.Where("(LOWER(p.name) LIKE ? OR LOWER(p.sku) LIKE ? OR p.barcode LIKE ?)", pattern, pattern, pattern)
		}
		if filter.CategoryID != nil {
			db = db.Where("p.category_id = ?", *filter.CategoryID)
		}
		if filter.Status != nil {
			db = db.Where("p.status = ?", string(*filter.Status))
		}
		return db
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []productRow
	if err := base().
		Select("p.*, c.name AS category_name").
		Order("p.name ASC, p.id ASC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]repository.ProductListItem, len(rows))
	for i := range rows {
		items[i] = repository.ProductListItem{
			Product:      mapper.ProductToDomain(&rows[i].ProductModel),
			CategoryName: rows[i].CategoryName,
		}
	}
	return items, total, nil
}

type summaryRow struct {
	ActiveProducts  int64
	InventoryValue  decimal.Decimal
	LowStock        int64
	WithoutMovement int64
}

func (r *ProductGormRepository) Summary(ctx context.Context, companyID uuid.UUID, noMovementSince time.Time) (repository.ProductSummary, error) {
	var row summaryRow
	err := r.db.GetDB(ctx).Model(&model.ProductModel{}).
		Select(`COUNT(*) AS active_products,
			COALESCE(SUM(stock * cost), 0) AS inventory_value,
			COUNT(*) FILTER (WHERE stock > 0 AND stock <= min_stock) AS low_stock,
			COUNT(*) FILTER (WHERE last_movement_at < ?) AS without_movement`, noMovementSince).
		Where("company_id = ? AND status = ?", companyID, string(entity.StatusActive)).
		Scan(&row).Error
	if err != nil {
		return repository.ProductSummary{}, err
	}
	return repository.ProductSummary{
		ActiveProducts:  row.ActiveProducts,
		InventoryValue:  row.InventoryValue,
		LowStock:        row.LowStock,
		WithoutMovement: row.WithoutMovement,
	}, nil
}
