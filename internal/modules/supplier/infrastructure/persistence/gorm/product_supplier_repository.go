package gorm

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
	"gorm.io/gorm"
	"strings"
)

type ProductSupplierGormRepository struct{ db *sharedDatabase.Database }

func NewProductSupplierGormRepository(db *sharedDatabase.Database) repository.ProductSupplierRepository {
	return &ProductSupplierGormRepository{db: db}
}

type productRow struct {
	model.ProductSupplierModel `gorm:"embedded"`
	Name                       string
	SKU                        *string
	EffectiveLeadTimeDays      *int
}

func (r *ProductSupplierGormRepository) List(ctx context.Context, companyID, supplierID uuid.UUID, search string, offset, limit int) ([]repository.ProductItem, int64, error) {
	var rows []productRow
	var total int64
	base := func(tx *gorm.DB) *gorm.DB {
		db := tx.Table("product_suppliers AS ps").Joins("JOIN products p ON p.id = ps.product_id AND p.company_id = ps.company_id").
			Joins("JOIN suppliers s ON s.id = ps.supplier_id AND s.company_id = ps.company_id AND s.deleted_at IS NULL").
			Where("ps.company_id = ? AND ps.supplier_id = ? AND p.deleted_at IS NULL", companyID, supplierID)
		if strings.TrimSpace(search) != "" {
			db = db.Where("(LOWER(p.name) LIKE ? OR LOWER(p.sku) LIKE ?)", pattern(search), pattern(search))
		}
		return db
	}
	err := r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := base(tx).Count(&total).Error; err != nil {
			return err
		}
		return base(tx).Select("ps.*, p.name, p.sku, COALESCE(ps.lead_time_days, s.lead_time_days) AS effective_lead_time_days").
			Order("p.name ASC, ps.product_id ASC").Offset(offset).Limit(limit).Scan(&rows).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	items := make([]repository.ProductItem, len(rows))
	for i := range rows {
		items[i] = repository.ProductItem{Association: mapper.AssociationToDomain(&rows[i].ProductSupplierModel), Name: rows[i].Name, SKU: rows[i].SKU, EffectiveLeadTimeDays: rows[i].EffectiveLeadTimeDays}
	}
	return items, total, nil
}
