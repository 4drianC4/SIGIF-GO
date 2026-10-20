package gorm

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	return translateDuplicate(db, db.Create(mapper.ProductToModel(product)).Error)
}

func (r *ProductGormRepository) Update(ctx context.Context, product *entity.Product) error {
	db := r.db.GetDB(ctx)
	return translateDuplicate(db, db.Save(mapper.ProductToModel(product)).Error)
}

func translateDuplicate(db *gorm.DB, err error) error {
	if err == nil || !isDuplicateKey(db, err) {
		return err
	}
	switch {
	case strings.Contains(err.Error(), "idx_products_company_name"):
		return service.ErrProductNameTaken
	case strings.Contains(err.Error(), "idx_products_company_barcode"):
		return service.ErrBarcodeTaken
	case strings.Contains(err.Error(), "idx_products_company_sku"):
		return service.ErrSKUTaken
	default:
		return err
	}
}

func (r *ProductGormRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Product, error) {
	var m model.ProductModel
	db := r.db.GetDB(ctx)
	if sharedDatabase.GetTx(ctx) != nil {
		// Inside a transaction the row is locked so two price changes cannot
		// record the same previous values in the history.
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := db.
		Where("company_id = ? AND id = ?", companyID, id).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ProductToDomain(&m), nil
}

func (r *ProductGormRepository) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(sharedDatabase.WithTx(ctx, tx))
	})
}

func (r *ProductGormRepository) SetStatus(ctx context.Context, companyID, id uuid.UUID, status entity.Status) error {
	result := r.db.GetDB(ctx).Model(&model.ProductModel{}).
		Where("company_id = ? AND id = ?", companyID, id).
		Update("status", string(status))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrProductNotFound
	}
	return nil
}

func (r *ProductGormRepository) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	result := r.db.GetDB(ctx).
		Where("company_id = ? AND id = ?", companyID, id).
		Delete(&model.ProductModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrProductNotFound
	}
	return nil
}

func (r *ProductGormRepository) FindByName(ctx context.Context, companyID uuid.UUID, name string, excludeID uuid.UUID) (*entity.Product, error) {
	return findHolder(r.db.GetDB(ctx), excludeID, "company_id = ? AND LOWER(name) = LOWER(?)", companyID, name)
}

func (r *ProductGormRepository) FindBySKU(ctx context.Context, companyID uuid.UUID, sku string, excludeID uuid.UUID) (*entity.Product, error) {
	return findHolder(r.db.GetDB(ctx).Unscoped(), excludeID, "company_id = ? AND sku = ?", companyID, sku)
}

func (r *ProductGormRepository) FindByBarcode(ctx context.Context, companyID uuid.UUID, barcode string, excludeID uuid.UUID) (*entity.Product, error) {
	return findHolder(r.db.GetDB(ctx).Unscoped(), excludeID, "company_id = ? AND barcode = ?", companyID, barcode)
}

// findHolder returns the product that matches condition, skipping excludeID
// and preferring a live product over a deleted one.
func findHolder(db *gorm.DB, excludeID uuid.UUID, condition string, args ...any) (*entity.Product, error) {
	query := db.Where(condition, args...)
	if excludeID != uuid.Nil {
		query = query.Where("id <> ?", excludeID)
	}
	// Find with a limit instead of First: not finding a holder is the usual
	// case and must not be logged as an error.
	var found []model.ProductModel
	if err := query.Order("deleted_at IS NOT NULL, created_at").Limit(1).Find(&found).Error; err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	return mapper.ProductToDomain(&found[0]), nil
}

// stockBalanceJoin aggregates the stock_balance cache per product so the list can
// report the physical stock. The effective stock falls back to products.stock
// while a product has no inventory rows (HU-04).
const stockBalanceJoin = "LEFT JOIN (SELECT company_id, product_id, SUM(quantity) AS total FROM stock_balance GROUP BY company_id, product_id) sb ON sb.company_id = p.company_id AND sb.product_id = p.id"

const effectiveStock = "COALESCE(sb.total, p.stock)"

type productRow struct {
	model.ProductModel `gorm:"embedded"`
	CategoryName       string
	EffectiveStock     decimal.Decimal
}

func (r *ProductGormRepository) List(ctx context.Context, companyID uuid.UUID, filter repository.ProductFilter, offset, limit int) ([]repository.ProductListItem, int64, error) {
	base := func() *gorm.DB {
		db := r.db.GetDB(ctx).Table("products AS p").
			Joins("JOIN categories c ON c.id = p.category_id").
			Joins(stockBalanceJoin).
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
		db = applyStockStatusFilter(db, filter.StockStatus)
		return db
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []productRow
	if err := base().
		Select("p.*, c.name AS category_name, " + effectiveStock + " AS effective_stock").
		Order("p.name ASC, p.id ASC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]repository.ProductListItem, len(rows))
	for i := range rows {
		product := mapper.ProductToDomain(&rows[i].ProductModel)
		product.Stock = rows[i].EffectiveStock
		items[i] = repository.ProductListItem{
			Product:      product,
			CategoryName: rows[i].CategoryName,
		}
	}
	return items, total, nil
}

func applyStockStatusFilter(db *gorm.DB, status *entity.StockStatus) *gorm.DB {
	if status == nil {
		return db
	}
	switch *status {
	case entity.StockStatusNormal:
		return db.Where(effectiveStock + " > p.min_stock")
	case entity.StockStatusLowStock:
		return db.Where(effectiveStock + " > 0 AND " + effectiveStock + " <= p.min_stock")
	case entity.StockStatusOutOfStock:
		return db.Where(effectiveStock + " <= 0")
	default:
		return db
	}
}

type summaryRow struct {
	ActiveProducts  int64
	InventoryValue  decimal.Decimal
	LowStock        int64
	WithoutMovement int64
}

func (r *ProductGormRepository) Summary(ctx context.Context, companyID uuid.UUID, noMovementSince time.Time) (repository.ProductSummary, error) {
	var row summaryRow
	err := r.db.GetDB(ctx).Table("products AS p").
		Joins(stockBalanceJoin).
		Select(`COUNT(*) AS active_products,
			COALESCE(SUM(`+effectiveStock+` * p.cost), 0) AS inventory_value,
			COUNT(*) FILTER (WHERE `+effectiveStock+` > 0 AND `+effectiveStock+` <= p.min_stock) AS low_stock,
			COUNT(*) FILTER (WHERE p.last_movement_at < ?) AS without_movement`, noMovementSince).
		Where("p.company_id = ? AND p.status = ?", companyID, string(entity.StatusActive)).
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
