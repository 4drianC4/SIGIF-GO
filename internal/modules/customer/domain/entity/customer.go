package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

// Customer es la entidad de dominio del cliente.
// El ID es generado por la base de datos (bigint autoincremental);
// al construir con NewCustomer se deja en 0 hasta que la BD lo asigne.
type Customer struct {
	ID             int64
	TenantID       uuid.UUID
	LegalName      string
	DocumentType   DocumentType
	DocumentNumber *string
	Phone          *string
	Email          *string
	Address        *string
	CreditLimit    decimal.Decimal
	CreditBalance  decimal.Decimal
	PointsAccrued  int
	Status         CustomerStatus
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

// NewCustomer crea un Customer con valores iniciales seguros.
//
// El formulario "Nuevo cliente" solo aporta legalName, documentType,
// documentNumber, phone y email.  Los demás campos financieros y de
// control son responsabilidad del backend.
//
// Nota: el modal no permite elegir tipo de documento; si documentType
// llega vacío se asigna DocumentTypeNationalID como valor por defecto.
func NewCustomer(
	clk clock.Clock,
	tenantID uuid.UUID,
	legalName string,
	documentType DocumentType,
	documentNumber *string,
	phone *string,
	email *string,
) *Customer {
	now := clk.NowUTC()

	if documentType == "" {
		documentType = DocumentTypeNationalID
	}

	// Normalizar campos opcionales de cadena: nil si viene vacío
	documentNumber = trimStringPtr(documentNumber)
	phone = trimStringPtr(phone)
	email = trimStringPtr(email)

	return &Customer{
		TenantID:       tenantID,
		LegalName:      strings.TrimSpace(legalName),
		DocumentType:   documentType,
		DocumentNumber: documentNumber,
		Phone:          phone,
		Email:          email,
		CreditLimit:    decimal.Zero,
		CreditBalance:  decimal.Zero,
		PointsAccrued:  0,
		Status:         CustomerStatusActive,
		CreatedAt:      now,
	}
}

// Update aplica los datos editables al cliente y registra la fecha de actualización.
func (c *Customer) Update(
	clk clock.Clock,
	legalName string,
	documentType DocumentType,
	documentNumber *string,
	phone *string,
	email *string,
	address *string,
) {
	now := clk.NowUTC()
	c.LegalName = strings.TrimSpace(legalName)
	c.DocumentType = documentType
	c.DocumentNumber = trimStringPtr(documentNumber)
	c.Phone = trimStringPtr(phone)
	c.Email = trimStringPtr(email)
	c.Address = trimStringPtr(address)
	c.UpdatedAt = &now
}

// Activate pone al cliente en estado activo.
func (c *Customer) Activate(clk clock.Clock) {
	now := clk.NowUTC()
	c.Status = CustomerStatusActive
	c.UpdatedAt = &now
}

// Deactivate pone al cliente en estado inactivo.
func (c *Customer) Deactivate(clk clock.Clock) {
	now := clk.NowUTC()
	c.Status = CustomerStatusInactive
	c.UpdatedAt = &now
}

// Block bloquea al cliente.
func (c *Customer) Block(clk clock.Clock) {
	now := clk.NowUTC()
	c.Status = CustomerStatusBlocked
	c.UpdatedAt = &now
}

// SoftDelete marca al cliente como eliminado sin borrarlo físicamente.
// Todas las lecturas normales deben excluir filas con DeletedAt distinto de nil.
func (c *Customer) SoftDelete(clk clock.Clock) {
	now := clk.NowUTC()
	c.DeletedAt = &now
	c.Status = CustomerStatusInactive
	c.UpdatedAt = &now
}

// IsDeleted informa si el cliente fue dado de baja lógicamente.
func (c *Customer) IsDeleted() bool {
	return c.DeletedAt != nil
}

// trimStringPtr retorna nil si el puntero apunta a una cadena vacía o solo espacios,
// o el puntero original (con el valor recortado) si tiene contenido real.
func trimStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
