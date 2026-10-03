package entity

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

var skuPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9._-]*$`)

type Product struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	CategoryID    uuid.UUID
	SKU           string
	Barcode       string
	Name          string
	Description   string
	UnitOfMeasure UnitOfMeasure
	CostPrice     decimal.Decimal
	SalePrice     decimal.Decimal
	Status        Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

type NewProductParams struct {
	TenantID      uuid.UUID
	CategoryID    uuid.UUID
	SKU           string
	Barcode       string
	Name          string
	Description   string
	UnitOfMeasure UnitOfMeasure
	CostPrice     decimal.Decimal
	SalePrice     decimal.Decimal
}

// NewProduct crea un producto activo aplicando las reglas de negocio de precios y SKU.
func NewProduct(clock clock.Clock, params NewProductParams) (*Product, error) {
	sku := NormalizeSKU(params.SKU)
	details := map[string]string{}

	if !skuPattern.MatchString(sku) {
		details["sku"] = "solo puede contener letras, números, '.', '-' y '_', y debe empezar por letra o número"
	}
	if params.CostPrice.IsNegative() {
		details["cost_price"] = "no puede ser negativo"
	} else if !hasAtMostTwoDecimals(params.CostPrice) {
		details["cost_price"] = "admite como máximo 2 decimales"
	}
	if !params.SalePrice.IsPositive() {
		details["sale_price"] = "debe ser mayor que 0"
	} else if !hasAtMostTwoDecimals(params.SalePrice) {
		details["sale_price"] = "admite como máximo 2 decimales"
	}
	if len(details) > 0 {
		return nil, sharedErrors.New(sharedErrors.CodeValidation, "validation failed", 400).WithDetails(details)
	}

	now := clock.NowUTC()
	return &Product{
		ID:            uuid.New(),
		TenantID:      params.TenantID,
		CategoryID:    params.CategoryID,
		SKU:           sku,
		Barcode:       strings.TrimSpace(params.Barcode),
		Name:          NormalizeName(params.Name),
		Description:   strings.TrimSpace(params.Description),
		UnitOfMeasure: params.UnitOfMeasure,
		CostPrice:     params.CostPrice,
		SalePrice:     params.SalePrice,
		Status:        StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// NormalizeSKU guarda el SKU sin espacios y en mayúsculas para que "abc-1" y "ABC-1" sean el mismo.
func NormalizeSKU(sku string) string {
	return strings.ToUpper(strings.TrimSpace(sku))
}

func hasAtMostTwoDecimals(d decimal.Decimal) bool {
	return d.Equal(d.Round(2))
}
