package entity

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"strings"
	"time"
	"unicode/utf8"
)

type ProductSupplier struct {
	ID           uuid.UUID
	CompanyID    uuid.UUID
	SupplierID   uuid.UUID
	ProductID    uuid.UUID
	SupplierCode *string
	LeadTimeDays *int
	CurrentCost  *decimal.Decimal
	Preferred    bool
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func (a *ProductSupplier) NormalizeAndValidate() error {
	if a.CompanyID == uuid.Nil || a.SupplierID == uuid.Nil || a.ProductID == uuid.Nil {
		return Invalid("association", "company, supplier and product are required")
	}
	if a.SupplierCode != nil {
		code := strings.TrimSpace(*a.SupplierCode)
		if utf8.RuneCountInString(code) > 60 {
			return Invalid("supplier_code", "maximum 60 characters")
		}
		if code == "" {
			a.SupplierCode = nil
		} else {
			a.SupplierCode = &code
		}
	}
	if err := ValidateLeadTime(a.LeadTimeDays); err != nil {
		return err
	}
	if a.CurrentCost != nil {
		return ValidateMoney("current_cost", *a.CurrentCost)
	}
	return nil
}
