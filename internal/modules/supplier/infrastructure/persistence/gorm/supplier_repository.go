package gorm

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/service"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func pattern(s string) string {
	return "%" + likeEscaper.Replace(strings.ToLower(strings.TrimSpace(s))) + "%"
}

type SupplierGormRepository struct{ db *sharedDatabase.Database }

func NewSupplierGormRepository(db *sharedDatabase.Database) repository.SupplierRepository {
	return &SupplierGormRepository{db: db}
}

func translateError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "idx_suppliers_company_tax_id":
			return service.ErrTaxIDTaken
		}
	}
	return err
}

func (r *SupplierGormRepository) Create(ctx context.Context, supplier *entity.Supplier) error {
	return translateError(r.db.GetDB(ctx).Create(mapper.ToModel(supplier)).Error)
}

func (r *SupplierGormRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Supplier, error) {
	db := r.db.GetDB(ctx).Where("company_id = ? AND id = ?", companyID, id)
	// Serialize supplier edits and deactivation.
	if sharedDatabase.GetTx(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var row model.SupplierModel
	err := db.First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mapper.ToDomain(&row), nil
}

func (r *SupplierGormRepository) Update(ctx context.Context, supplier *entity.Supplier) error {
	m := mapper.ToModel(supplier)
	result := r.db.GetDB(ctx).Model(m).Where("company_id = ?", m.CompanyID).
		Select("trade_name", "legal_name", "tax_id", "contact", "phone", "email", "address",
			"main_industry", "payment_terms", "lead_time_days", "minimum_order", "updated_at").Updates(m)
	if result.Error != nil {
		return translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return service.ErrSupplierNotFound
	}
	return nil
}

func (r *SupplierGormRepository) Deactivate(ctx context.Context, companyID, id uuid.UUID, at time.Time) error {
	result := r.db.GetDB(ctx).Model(&model.SupplierModel{}).Where("company_id = ? AND id = ?", companyID, id).
		Updates(map[string]any{"status": string(entity.StatusInactive), "updated_at": at})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrSupplierNotFound
	}
	return nil
}

func (r *SupplierGormRepository) ExistsByTaxID(ctx context.Context, companyID uuid.UUID, taxID string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.SupplierModel{}).
		Where("company_id = ? AND tax_id = ? AND id <> ?", companyID, taxID, excludeID).Count(&count).Error
	return count > 0, err
}

type supplierListRow struct {
	model.SupplierModel `gorm:"embedded"`
	ProductsCount       int64
}

func (r *SupplierGormRepository) List(ctx context.Context, companyID uuid.UUID, filter repository.ListFilter, offset, limit int) ([]repository.SupplierListItem, int64, error) {
	var rows []supplierListRow
	var total int64
	base := func(db *gorm.DB) *gorm.DB {
		db = db.Table("suppliers AS s").Where("s.company_id = ? AND s.deleted_at IS NULL", companyID)
		if strings.TrimSpace(filter.Search) != "" {
			p := pattern(filter.Search)
			db = db.Where("(LOWER(s.legal_name) LIKE ? OR LOWER(s.trade_name) LIKE ? OR s.tax_id LIKE ?)", p, p, p)
		}
		if filter.Status != nil {
			db = db.Where("s.status = ?", string(*filter.Status))
		}
		if filter.MainIndustry != "" {
			db = db.Where("s.main_industry = ?", filter.MainIndustry)
		}
		return db
	}
	column := map[string]string{"name": "COALESCE(s.trade_name, s.legal_name)", "legal_name": "s.legal_name", "created_at": "s.created_at", "lead_time_days": "s.lead_time_days"}[filter.SortBy]
	if column == "" {
		column = "COALESCE(s.trade_name, s.legal_name)"
	}
	direction := "ASC"
	if filter.SortOrder == "desc" {
		direction = "DESC"
	}
	err := r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := base(tx).Count(&total).Error; err != nil {
			return err
		}
		return base(tx).Select(`s.*,
			(SELECT count(*) FROM product_suppliers ps JOIN products p ON p.id = ps.product_id AND p.company_id = ps.company_id
			 WHERE ps.company_id = s.company_id AND ps.supplier_id = s.id AND p.deleted_at IS NULL) AS products_count`).
			Order(column + " " + direction + " NULLS LAST, s.id ASC").Offset(offset).Limit(limit).Scan(&rows).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	items := make([]repository.SupplierListItem, len(rows))
	for i := range rows {
		items[i] = repository.SupplierListItem{Supplier: mapper.ToDomain(&rows[i].SupplierModel), ProductsCount: rows[i].ProductsCount}
	}
	return items, total, nil
}

func (r *SupplierGormRepository) Summary(ctx context.Context, companyID uuid.UUID) (repository.SupplierSummary, error) {
	var summary repository.SupplierSummary
	err := r.db.GetDB(ctx).Model(&model.SupplierModel{}).
		Where("company_id = ? AND status = 'active'", companyID).Count(&summary.ActiveSuppliers).Error
	return summary, err
}

func (r *SupplierGormRepository) FormOptions(ctx context.Context, companyID uuid.UUID) (repository.FormOptions, error) {
	options := repository.FormOptions{Industries: []string{}, PaymentTerms: []string{}}
	for _, dest := range []struct {
		column string
		values *[]string
	}{
		{"main_industry", &options.Industries}, {"payment_terms", &options.PaymentTerms},
	} {
		if err := r.db.GetDB(ctx).Model(&model.SupplierModel{}).Where("company_id = ?", companyID).
			Where(dest.column+" IS NOT NULL").Distinct(dest.column).Order(dest.column).Pluck(dest.column, dest.values).Error; err != nil {
			return options, err
		}
	}
	return options, nil
}

func (r *SupplierGormRepository) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error { return fn(sharedDatabase.WithTx(ctx, tx)) })
}
