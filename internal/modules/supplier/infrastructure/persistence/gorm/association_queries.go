package gorm

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/model"
)

type supplierAssociationRow struct {
	model.SupplierModel   `gorm:"embedded"`
	Association           model.ProductSupplierModel `gorm:"embedded;embeddedPrefix:association_"`
	EffectiveLeadTimeDays *int
}

func (r *ProductSupplierGormRepository) ListSuppliers(ctx context.Context, companyID, productID uuid.UUID, search string, offset, limit int) ([]repository.SupplierItem, int64, error) {
	var rows []supplierAssociationRow
	var total int64
	base := func(tx *gorm.DB) *gorm.DB {
		db := tx.Table("product_suppliers ps").Joins("JOIN suppliers s ON s.id = ps.supplier_id AND s.company_id = ps.company_id").
			Where("ps.company_id = ? AND ps.product_id = ? AND s.deleted_at IS NULL", companyID, productID)
		if strings.TrimSpace(search) != "" {
			db = db.Where("(LOWER(s.legal_name) LIKE ? OR LOWER(s.trade_name) LIKE ? OR s.tax_id LIKE ? OR LOWER(ps.supplier_code) LIKE ?)", pattern(search), pattern(search), pattern(search), pattern(search))
		}
		return db
	}
	err := r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := base(tx).Count(&total).Error; err != nil {
			return err
		}
		return base(tx).Select(`s.*, ps.id AS association_id, ps.company_id AS association_company_id,
		 ps.supplier_id AS association_supplier_id, ps.product_id AS association_product_id,
		 ps.supplier_code AS association_supplier_code, ps.current_cost AS association_current_cost,
		 ps.lead_time_days AS association_lead_time_days, ps.preferred AS association_preferred,
		 ps.created_at AS association_created_at, ps.updated_at AS association_updated_at,
		 COALESCE(ps.lead_time_days, s.lead_time_days) AS effective_lead_time_days`).
			Order("ps.preferred DESC, COALESCE(s.trade_name, s.legal_name) ASC, s.id ASC").Offset(offset).Limit(limit).Scan(&rows).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	items := make([]repository.SupplierItem, len(rows))
	for i := range rows {
		items[i] = repository.SupplierItem{Association: mapper.AssociationToDomain(&rows[i].Association), Supplier: mapper.ToDomain(&rows[i].SupplierModel), EffectiveLeadTimeDays: rows[i].EffectiveLeadTimeDays}
	}
	return items, total, nil
}

func (r *ProductSupplierGormRepository) ProductOptions(ctx context.Context, companyID, supplierID uuid.UUID, filter repository.ProductOptionFilter, offset, limit int) ([]repository.ProductOption, int64, error) {
	items := []repository.ProductOption{}
	var total int64
	base := func(tx *gorm.DB) *gorm.DB {
		db := tx.Table("products p").
			Joins("JOIN categories c ON c.id = p.category_id AND c.company_id = p.company_id AND c.deleted_at IS NULL").
			Joins("JOIN units_of_measure u ON u.id = p.unit_of_measure_id").
			Joins("LEFT JOIN product_suppliers ps ON ps.product_id = p.id AND ps.company_id = p.company_id AND ps.supplier_id = ?", supplierID).
			Where("p.company_id = ? AND p.deleted_at IS NULL", companyID)
		if strings.TrimSpace(filter.Search) != "" {
			db = db.Where("(LOWER(p.name) LIKE ? OR LOWER(p.sku) LIKE ? OR p.barcode LIKE ?)", pattern(filter.Search), pattern(filter.Search), pattern(filter.Search))
		}
		if filter.CategoryID != nil {
			db = db.Where(`p.category_id IN (WITH RECURSIVE selected AS (
			 SELECT id FROM categories WHERE id = ? AND company_id = ? AND deleted_at IS NULL
			 UNION SELECT c.id FROM categories c JOIN selected s ON c.parent_id = s.id WHERE c.company_id = ? AND c.deleted_at IS NULL
			) SELECT id FROM selected)`, *filter.CategoryID, companyID, companyID)
		}
		return db
	}
	err := r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := base(tx).Count(&total).Error; err != nil {
			return err
		}
		if err := base(tx).Select("p.id, p.name, p.sku, p.barcode, p.status, p.category_id, c.name AS category_name, p.stock, u.abbreviation AS unit_abbreviation, (ps.id IS NOT NULL) AS is_associated").
			Order("p.name ASC, p.id ASC").Offset(offset).Limit(limit).Scan(&items).Error; err != nil {
			return err
		}
		return fillCategoryPaths(tx, companyID, items)
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return items, total, err
}

// Resolve the breadcrumb from tenant-scoped categories, including deeper trees.
func fillCategoryPaths(tx *gorm.DB, companyID uuid.UUID, items []repository.ProductOption) error {
	if len(items) == 0 {
		return nil
	}
	type category struct {
		ID       uuid.UUID
		ParentID *uuid.UUID
		Name     string
	}
	var rows []category
	if err := tx.Table("categories").Select("id, parent_id, name").Where("company_id = ? AND deleted_at IS NULL", companyID).Scan(&rows).Error; err != nil {
		return err
	}
	byID := make(map[uuid.UUID]category, len(rows))
	for _, c := range rows {
		byID[c.ID] = c
	}
	for i := range items {
		names, seen := []string{}, map[uuid.UUID]bool{}
		for id := items[i].CategoryID; !seen[id]; {
			c, ok := byID[id]
			if !ok {
				break
			}
			seen[id] = true
			names = append(names, c.Name)
			if c.ParentID == nil {
				break
			}
			id = *c.ParentID
		}
		for a, b := 0, len(names)-1; a < b; a, b = a+1, b-1 {
			names[a], names[b] = names[b], names[a]
		}
		items[i].CategoryPath = strings.Join(names, " / ")
	}
	return nil
}

type supplierOptionRow struct {
	model.SupplierModel `gorm:"embedded"`
	IsAssociated        bool
}

func (r *ProductSupplierGormRepository) SupplierOptions(ctx context.Context, companyID, productID uuid.UUID, search string, offset, limit int) ([]repository.SupplierOption, int64, error) {
	var rows []supplierOptionRow
	var total int64
	base := func(tx *gorm.DB) *gorm.DB {
		db := tx.Table("suppliers s").
			Joins("LEFT JOIN product_suppliers ps ON ps.supplier_id = s.id AND ps.company_id = s.company_id AND ps.product_id = ?", productID).
			Where("s.company_id = ? AND s.deleted_at IS NULL", companyID)
		if strings.TrimSpace(search) != "" {
			db = db.Where("(LOWER(s.legal_name) LIKE ? OR LOWER(s.trade_name) LIKE ? OR s.tax_id LIKE ?)", pattern(search), pattern(search), pattern(search))
		}
		return db
	}
	err := r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := base(tx).Count(&total).Error; err != nil {
			return err
		}
		return base(tx).Select("s.*, (ps.id IS NOT NULL) AS is_associated").Order("COALESCE(s.trade_name, s.legal_name) ASC, s.id ASC").Offset(offset).Limit(limit).Scan(&rows).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	items := make([]repository.SupplierOption, len(rows))
	for i := range rows {
		items[i] = repository.SupplierOption{Supplier: mapper.ToDomain(&rows[i].SupplierModel), IsAssociated: rows[i].IsAssociated}
	}
	return items, total, nil
}
