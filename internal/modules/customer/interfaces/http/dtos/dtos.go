package dtos

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

type CreateCustomerRequest struct {
	LegalName      string  `json:"legal_name" validate:"required,min=1,max=160"`
	DocumentType   string  `json:"document_type" validate:"omitempty,oneof=national_id tax_id passport other"`
	DocumentNumber *string `json:"document_number" validate:"omitempty,max=30"`
	Phone          *string `json:"phone" validate:"omitempty,max=30"`
	Email          *string `json:"email" validate:"omitempty,email,max=160"`
}

type UpdateCustomerRequest struct {
	LegalName      string  `json:"legal_name" validate:"required,min=1,max=160"`
	DocumentType   string  `json:"document_type" validate:"omitempty,oneof=national_id tax_id passport other"`
	DocumentNumber *string `json:"document_number" validate:"omitempty,max=30"`
	Phone          *string `json:"phone" validate:"omitempty,max=30"`
	Email          *string `json:"email" validate:"omitempty,email,max=160"`
	Address        *string `json:"address" validate:"omitempty,max=200"`
}

type ChangeStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive blocked"`
}

type CustomerListResponse struct {
	Customers []dto.Customer `json:"customers"`
	Total     int64          `json:"total"`
	Page      int            `json:"page"`
	Limit     int            `json:"limit"`
}

func ToResponse(c dto.Customer) dto.Customer {
	return c
}

func ToResponseList(customers []dto.Customer) []dto.Customer {
	return customers
}

func TenantIDFromContext(c *fiber.Ctx) uuid.UUID {
    if id, ok := c.Locals("company_id").(uuid.UUID); ok {
        return id
    }
    return uuid.Nil
}

func DocumentTypeFromRequest(dt string) entity.DocumentType {
	if dt == "" {
		return entity.DocumentTypeNationalID
	}
	return entity.DocumentType(dt)
}

func StatusFromRequest(s string) entity.CustomerStatus {
	return entity.CustomerStatus(s)
}
