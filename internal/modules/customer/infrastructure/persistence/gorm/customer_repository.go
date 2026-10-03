package gorm

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// CustomerGormRepository implementa repository.CustomerRepository usando GORM.
type CustomerGormRepository struct {
	db *sharedDatabase.Database
}

func NewCustomerGormRepository(db *sharedDatabase.Database) repository.CustomerRepository {
	return &CustomerGormRepository{db: db}
}

// Create persiste un nuevo cliente y copia el ID generado por la BD a la entidad.
func (r *CustomerGormRepository) Create(ctx context.Context, customer *entity.Customer) error {
	m := mapper.ToModel(customer)
	if err := r.db.GetDB(ctx).Create(m).Error; err != nil {
		if isUniqueViolation(err) {
			return sharedErrors.New(sharedErrors.CodeConflict,
				"a customer with this document already exists in this tenant", 409)
		}
		return err
	}
	// Retroalimentar el ID generado a la entidad de dominio
	customer.ID = m.ID
	return nil
}

// GetByID retorna el cliente filtrando por tenant y excluyendo soft-deleted.
// Retorna (nil, nil) si no existe o el tenant no coincide.
func (r *CustomerGormRepository) GetByID(ctx context.Context, tenantID uuid.UUID, id int64) (*entity.Customer, error) {
	var m model.CustomerModel
	err := r.db.GetDB(ctx).
		Where("customer_id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomain(&m), nil
}

// List retorna una página de clientes del tenant aplicando filtros dinámicos.
// El conteo se realiza antes del offset/limit para que la paginación sea coherente.
func (r *CustomerGormRepository) List(
	ctx context.Context,
	tenantID uuid.UUID,
	filter repository.ListFilter,
	offset, limit int,
) ([]*entity.Customer, int64, error) {
	var models []model.CustomerModel
	var total int64

	db := r.db.GetDB(ctx).Model(&model.CustomerModel{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// Filtro de búsqueda libre: coincidencia parcial en múltiples columnas
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := "%" + strings.ToLower(q) + "%"
		db = db.Where(
			"LOWER(legal_name) LIKE ? OR LOWER(document_number) LIKE ? OR LOWER(phone) LIKE ? OR LOWER(email) LIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}

	// Filtro de estado
	if filter.Status != nil {
		db = db.Where("status = ?", string(*filter.Status))
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.
		Order("created_at DESC, customer_id DESC").
		Offset(offset).
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	customers := make([]*entity.Customer, len(models))
	for i := range models {
		customers[i] = mapper.ToDomain(&models[i])
	}

	return customers, total, nil
}

// ExistsByDocument verifica unicidad dentro del tenant para clientes activos.
func (r *CustomerGormRepository) ExistsByDocument(
	ctx context.Context,
	tenantID uuid.UUID,
	docType entity.DocumentType,
	docNumber string,
) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Model(&model.CustomerModel{}).
		Where("tenant_id = ? AND document_type = ? AND document_number = ? AND deleted_at IS NULL",
			tenantID, string(docType), docNumber).
		Count(&count).Error
	return count > 0, err
}

// Update persiste los cambios de un cliente existente.
func (r *CustomerGormRepository) Update(ctx context.Context, customer *entity.Customer) error {
	m := mapper.ToModel(customer)
	if err := r.db.GetDB(ctx).Save(m).Error; err != nil {
		if isUniqueViolation(err) {
			return sharedErrors.New(sharedErrors.CodeConflict,
				"a customer with this document already exists in this tenant", 409)
		}
		return err
	}
	return nil
}

// isUniqueViolation detecta violaciones de clave única de PostgreSQL (código 23505).
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
