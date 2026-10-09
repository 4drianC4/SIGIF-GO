package gorm

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
	"github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type CustomerGormRepository struct {
	db *sharedDatabase.Database
}

func NewCustomerGormRepository(db *sharedDatabase.Database) repository.CustomerRepository {
	return &CustomerGormRepository{db: db}
}

func (r *CustomerGormRepository) Create(ctx context.Context, customer *entity.Customer) error {
	m := mapper.ToModel(customer)
	if err := r.db.GetDB(ctx).Create(m).Error; err != nil {
		if isUniqueViolation(err) {
			return service.ErrDocumentTaken
		}
		return err
	}
	customer.ID = m.ID
	return nil
}

func (r *CustomerGormRepository) GetByID(ctx context.Context, companyID, id uuid.UUID) (*entity.Customer, error) {
	var m model.CustomerModel
	err := r.db.GetDB(ctx).
		Where("customer_id = ? AND company_id = ? AND deleted_at IS NULL", id, companyID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomain(&m), nil
}

// sortColumns whitelists the columns a listing can be ordered by.
var sortColumns = map[repository.SortField]string{
	repository.SortByLegalName: "legal_name",
	repository.SortByCreatedAt: "created_at",
}

// List runs the count and the page inside a single read-only, repeatable-read
// transaction so both see the same snapshot and no write can happen.
func (r *CustomerGormRepository) List(
	ctx context.Context,
	companyID uuid.UUID,
	filter repository.ListFilter,
	offset, limit int,
) ([]*entity.Customer, int64, error) {
	var models []model.CustomerModel
	var total int64

	scope := func(db *gorm.DB) *gorm.DB {
		db = db.Where("company_id = ? AND deleted_at IS NULL", companyID)

		if q := strings.TrimSpace(filter.Q); q != "" {
			pattern := "%" + escapeLike(strings.ToLower(q)) + "%"
			db = db.Where(
				"LOWER(legal_name) LIKE ? OR LOWER(document_number) LIKE ? OR LOWER(phone) LIKE ? OR LOWER(email) LIKE ?",
				pattern, pattern, pattern, pattern,
			)
		}

		if filter.Status != nil {
			db = db.Where("status = ?", string(*filter.Status))
		}
		return db
	}

	column, ok := sortColumns[filter.SortBy]
	if !ok {
		column = sortColumns[repository.SortByCreatedAt]
	}
	direction := "DESC"
	if filter.SortOrder == repository.SortAsc {
		direction = "ASC"
	}

	err := r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.CustomerModel{}).Scopes(scope).Count(&total).Error; err != nil {
			return err
		}
		return tx.Model(&model.CustomerModel{}).Scopes(scope).
			Order(column + " " + direction + ", customer_id " + direction).
			Offset(offset).
			Limit(limit).
			Find(&models).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}

	customers := make([]*entity.Customer, len(models))
	for i := range models {
		customers[i] = mapper.ToDomain(&models[i])
	}

	return customers, total, nil
}

func (r *CustomerGormRepository) ExistsByDocument(
	ctx context.Context,
	companyID uuid.UUID,
	docType entity.DocumentType,
	docNumber string,
) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.CustomerModel{}).
		Where("company_id = ? AND document_type = ? AND document_number = ? AND deleted_at IS NULL",
			companyID, string(docType), docNumber).
		Count(&count).Error
	return count > 0, err
}

// Update writes every column except the keys and created_at, only on the
// customer's own non-deleted row. Unlike Save it never falls back to INSERT.
func (r *CustomerGormRepository) Update(ctx context.Context, customer *entity.Customer) error {
	m := mapper.ToModel(customer)
	result := r.db.GetDB(ctx).Model(m).
		Where("company_id = ? AND deleted_at IS NULL", m.CompanyID).
		Select("*").
		Omit("customer_id", "company_id", "created_at").
		Updates(m)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return service.ErrDocumentTaken
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrCustomerNotFound
	}
	return nil
}

func (r *CustomerGormRepository) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(sharedDatabase.WithTx(ctx, tx))
	})
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}

// escapeLike makes user input match literally inside a LIKE pattern.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(s)
}
