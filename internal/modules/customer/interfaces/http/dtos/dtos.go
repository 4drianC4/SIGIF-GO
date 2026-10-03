package dtos

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

// CreateCustomerRequest representa el payload para crear un cliente.
// DocumentType es opcional y se asume 'national_id' por defecto.
type CreateCustomerRequest struct {
	LegalName      string  `json:"legal_name" validate:"required,min=1,max=160"`
	DocumentType   string  `json:"document_type" validate:"omitempty,oneof=national_id tax_id passport other"`
	DocumentNumber *string `json:"document_number" validate:"omitempty,max=30"`
	Phone          *string `json:"phone" validate:"omitempty,max=30"`
	Email          *string `json:"email" validate:"omitempty,email,max=160"`
}

// UpdateCustomerRequest representa el payload para actualizar un cliente.
type UpdateCustomerRequest struct {
	LegalName      string  `json:"legal_name" validate:"required,min=1,max=160"`
	DocumentType   string  `json:"document_type" validate:"omitempty,oneof=national_id tax_id passport other"`
	DocumentNumber *string `json:"document_number" validate:"omitempty,max=30"`
	Phone          *string `json:"phone" validate:"omitempty,max=30"`
	Email          *string `json:"email" validate:"omitempty,email,max=160"`
	Address        *string `json:"address" validate:"omitempty,max=200"`
}

// ChangeStatusRequest representa el payload para cambiar el estado.
type ChangeStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive blocked"`
}

// CustomerListResponse representa la respuesta paginada de clientes.
type CustomerListResponse struct {
	Customers []dto.Customer `json:"customers"`
	Total     int64          `json:"total"`
	Page      int            `json:"page"`
	Limit     int            `json:"limit"`
}

// ToResponse convierte un DTO plano a la estructura de respuesta final.
// En este caso, como el DTO de aplicación ya tiene los JSON tags correctos,
// se devuelve directamente. Podría envolverse si fuera necesario.
func ToResponse(c dto.Customer) dto.Customer {
	return c
}

// ToResponseList convierte una lista de DTOs planos.
func ToResponseList(customers []dto.Customer) []dto.Customer {
	return customers
}

// TenantIDFromContext extrae el tenant_id del contexto autenticado de Fiber.
func TenantIDFromContext(c *fiber.Ctx) uuid.UUID {
	v := c.Locals("tenant_id")
	if v == nil {
		return uuid.Nil
	}
	if id, ok := v.(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// DocumentTypeFromRequest mapea el string del request al enum del dominio.
func DocumentTypeFromRequest(dt string) entity.DocumentType {
	if dt == "" {
		return entity.DocumentTypeNationalID // Por defecto
	}
	return entity.DocumentType(dt)
}

// StatusFromRequest mapea el string del request al enum del dominio.
func StatusFromRequest(s string) entity.CustomerStatus {
	return entity.CustomerStatus(s)
}
