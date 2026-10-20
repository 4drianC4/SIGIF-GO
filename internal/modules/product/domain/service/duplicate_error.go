package service

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// Codes of the values a product cannot share with another product of the
// company (HU-06-01).
const (
	CodeNameDuplicate    = "NAME_DUPLICATE"
	CodeSKUDuplicate     = "SKU_DUPLICATE"
	CodeBarcodeDuplicate = "BARCODE_DUPLICATE"
)

// ExistingProduct identifies the product that already holds a value. A
// deleted product still holds its SKU and barcode.
type ExistingProduct struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	SKU     *string       `json:"sku"`
	Barcode *string       `json:"barcode"`
	Status  entity.Status `json:"status"`
	Deleted bool          `json:"deleted"`
}

// DuplicateConflict is one duplicated value of a 409 response.
type DuplicateConflict struct {
	Field           string          `json:"field"`
	Code            string          `json:"code"`
	Value           string          `json:"value"`
	ExistingProduct ExistingProduct `json:"existing_product"`
}

// DuplicateDetails is the details object of a 409 caused by duplicated values.
type DuplicateDetails struct {
	Conflicts []DuplicateConflict `json:"conflicts"`
}

var conflictCodes = map[string]string{
	"name":    CodeNameDuplicate,
	"sku":     CodeSKUDuplicate,
	"barcode": CodeBarcodeDuplicate,
}

var fieldErrors = map[string]*sharedErrors.AppError{
	"name":    ErrProductNameTaken,
	"sku":     ErrSKUTaken,
	"barcode": ErrBarcodeTaken,
}

// newDuplicateConflict describes value of field held by the existing product.
func newDuplicateConflict(field, value string, existing *entity.Product) DuplicateConflict {
	return DuplicateConflict{
		Field: field,
		Code:  conflictCodes[field],
		Value: value,
		ExistingProduct: ExistingProduct{
			ID:      existing.ID.String(),
			Name:    existing.Name,
			SKU:     optionalString(existing.SKU),
			Barcode: optionalString(existing.Barcode),
			Status:  existing.Status,
			Deleted: existing.DeletedAt != nil,
		},
	}
}

// newDuplicateError builds the 409 CONFLICT that lists every duplicated
// value. Its cause is the error of the first field, so errors.Is keeps
// matching ErrSKUTaken, ErrBarcodeTaken or ErrProductNameTaken.
func newDuplicateError(conflicts []DuplicateConflict) error {
	cause := fieldErrors[conflicts[0].Field]
	message := cause.Message
	if len(conflicts) > 1 {
		message = "product has duplicated values"
	}
	return &sharedErrors.AppError{
		Code:       sharedErrors.CodeConflict,
		Message:    message,
		StatusCode: 409,
		Details:    DuplicateDetails{Conflicts: conflicts},
		Cause:      cause,
	}
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
