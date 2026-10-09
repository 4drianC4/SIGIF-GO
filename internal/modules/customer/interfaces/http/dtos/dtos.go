package dtos

import (
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

// PatchCustomerRequest is the partial body of PATCH /customers/:id. A nil
// field was not sent and keeps its current value.
type PatchCustomerRequest struct {
	LegalName      *string `json:"legal_name" validate:"omitempty,min=1,max=160"`
	DocumentType   *string `json:"document_type" validate:"omitempty,oneof=national_id tax_id passport other"`
	DocumentNumber *string `json:"document_number" validate:"omitempty,max=30"`
	Phone          *string `json:"phone" validate:"omitempty,max=30"`
	Email          *string `json:"email" validate:"omitempty,email,max=160"`
	Address        *string `json:"address" validate:"omitempty,max=200"`
}

// PatchEditableFields lists the body keys PATCH accepts. Optional ones can be
// cleared with null or an empty string.
var PatchEditableFields = map[string]bool{
	"legal_name":      false,
	"document_type":   false,
	"document_number": true,
	"phone":           true,
	"email":           true,
	"address":         true,
}

// PatchReadOnlyFields are customer attributes that exist but cannot be changed
// through PATCH; any other unknown key is reported as unknown.
var PatchReadOnlyFields = map[string]bool{
	"id":             true,
	"customer_id":    true,
	"company_id":     true,
	"status":         true,
	"credit_limit":   true,
	"credit_balance": true,
	"points_accrued": true,
	"created_at":     true,
	"updated_at":     true,
	"deleted_at":     true,
}

type ChangeStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive blocked"`
}

// StatusAll lists customers in every status.
const StatusAll = "all"

// ListCustomersRequest holds the query string of GET /customers. Page and
// limit are pointers so an explicit 0 is rejected instead of defaulted.
type ListCustomersRequest struct {
	Q         string `json:"q" query:"q" validate:"max=100"`
	Status    string `json:"status" query:"status" validate:"omitempty,oneof=active inactive blocked all"`
	Page      *int   `json:"page" query:"page" validate:"omitempty,min=1"`
	Limit     *int   `json:"limit" query:"limit" validate:"omitempty,min=1"`
	SortBy    string `json:"sort_by" query:"sort_by" validate:"omitempty,oneof=legal_name created_at"`
	SortOrder string `json:"sort_order" query:"sort_order" validate:"omitempty,oneof=asc desc"`
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
