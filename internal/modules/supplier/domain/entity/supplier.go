package entity

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type Supplier struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Fields
	Status    Status
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// Fields contains only the editable supplier master data.
type Fields struct {
	TradeName    *string
	LegalName    string
	TaxID        *string
	Contact      *string
	Phone        *string
	Email        *string
	Address      *string
	MainIndustry *string
	PaymentTerms *string
	LeadTimeDays *int
	MinimumOrder *decimal.Decimal
}

func (f *Fields) Normalize() {
	f.LegalName = strings.TrimSpace(f.LegalName)
	for _, p := range []**string{&f.TradeName, &f.TaxID, &f.Contact, &f.Phone, &f.Email, &f.Address, &f.MainIndustry, &f.PaymentTerms} {
		if *p != nil {
			v := strings.TrimSpace(**p)
			if v == "" {
				*p = nil
			} else {
				*p = &v
			}
		}
	}
}

func Invalid(field, message string) error {
	return sharedErrors.New(sharedErrors.CodeValidation, "validation failed", 400).
		WithDetails(map[string]string{field: message})
}

func (f Fields) Validate() error {
	if f.LegalName == "" || utf8.RuneCountInString(f.LegalName) > 160 {
		return Invalid("legal_name", "required; maximum 160 characters")
	}
	for _, v := range []struct {
		name  string
		value *string
		max   int
	}{
		{"trade_name", f.TradeName, 160}, {"tax_id", f.TaxID, 20},
		{"contact", f.Contact, 120}, {"phone", f.Phone, 30},
		{"email", f.Email, 160}, {"address", f.Address, 200},
		{"main_industry", f.MainIndustry, 80}, {"payment_terms", f.PaymentTerms, 60},
	} {
		if v.value != nil && utf8.RuneCountInString(*v.value) > v.max {
			return Invalid(v.name, "maximum length exceeded")
		}
	}
	if f.TaxID != nil {
		for _, c := range *f.TaxID {
			if c < '0' || c > '9' {
				return Invalid("tax_id", "must contain only digits")
			}
		}
	}
	if f.Email != nil {
		parsed, err := mail.ParseAddress(*f.Email)
		if err != nil || parsed.Address != *f.Email {
			return Invalid("email", "invalid email address")
		}
	}
	if err := ValidateLeadTime(f.LeadTimeDays); err != nil {
		return err
	}
	if f.MinimumOrder != nil {
		return ValidateMoney("minimum_order", *f.MinimumOrder)
	}
	return nil
}

func ValidateLeadTime(days *int) error {
	if days != nil && (*days < 0 || *days > 32767) {
		return Invalid("lead_time_days", "must be between 0 and 32767")
	}
	return nil
}

func ValidateMoney(field string, value decimal.Decimal) error {
	if value.IsNegative() || value.GreaterThan(decimal.New(999999999999, -2)) || !value.Equal(value.Round(2)) {
		return Invalid(field, "must be between 0 and 9999999999.99 with at most two decimal places")
	}
	return nil
}

func NewSupplier(clk clock.Clock, companyID uuid.UUID, fields Fields) (*Supplier, error) {
	fields.Normalize()
	if companyID == uuid.Nil {
		return nil, Invalid("company_id", "required")
	}
	if err := fields.Validate(); err != nil {
		return nil, err
	}
	return &Supplier{ID: uuid.New(), CompanyID: companyID, Fields: fields, Status: StatusActive, CreatedAt: clk.NowUTC()}, nil
}

func (s *Supplier) Name() string {
	if s.TradeName != nil {
		return *s.TradeName
	}
	return s.LegalName
}
