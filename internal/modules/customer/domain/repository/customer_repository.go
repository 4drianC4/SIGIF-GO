package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

// ListFilter agrupa los filtros opcionales para el listado de clientes.
// Todos los filtros se aplican siempre dentro del tenant autenticado.
type ListFilter struct {
	// Q busca coincidencias parciales en legal_name, document_number, phone y email.
	Q string
	// Status, cuando no es nil, restringe los resultados al estado indicado.
	Status *entity.CustomerStatus
}

// Reader define las operaciones de lectura sobre clientes.
type Reader interface {
	// GetByID retorna el cliente con ese ID dentro del tenant.
	// Retorna (nil, nil) si no existe o pertenece a otro tenant.
	GetByID(ctx context.Context, tenantID uuid.UUID, id int64) (*entity.Customer, error)

	// List retorna una página de clientes que cumplan los filtros.
	// offset y limit controlan la paginación; siempre excluye registros con soft-delete.
	List(ctx context.Context, tenantID uuid.UUID, filter ListFilter, offset, limit int) ([]*entity.Customer, int64, error)

	// ExistsByDocument verifica si ya existe un cliente activo en el mismo tenant
	// con el mismo tipo y número de documento.
	ExistsByDocument(ctx context.Context, tenantID uuid.UUID, docType entity.DocumentType, docNumber string) (bool, error)
}

// Writer define las operaciones de escritura sobre clientes.
type Writer interface {
	// Create persiste un nuevo cliente y actualiza su ID con el valor generado por la BD.
	Create(ctx context.Context, customer *entity.Customer) error

	// Update persiste los cambios sobre un cliente existente.
	Update(ctx context.Context, customer *entity.Customer) error
}

// CustomerRepository es el puerto completo de persistencia de clientes.
type CustomerRepository interface {
	Reader
	Writer
}
