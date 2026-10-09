package entity

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

var skuPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9._-]*$`)

var hundred = decimal.NewFromInt(100)

type StockStatus string

const (
	StockStatusNormal     StockStatus = "normal"
	StockStatusLowStock   StockStatus = "low_stock"
	StockStatusOutOfStock StockStatus = "out_of_stock"
)

type Product struct {
	ID              uuid.UUID
	CompanyID       uuid.UUID
	CategoryID      uuid.UUID
	UnitOfMeasureID uuid.UUID
	SKU             string
	Barcode         string
	Name            string
	Description     string
	Cost            decimal.Decimal
	SalePrice       decimal.Decimal
	Stock           decimal.Decimal
	MinStock        decimal.Decimal
	Status          Status
	LastMovementAt  time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type NewProductParams struct {
	CompanyID       uuid.UUID
	CategoryID      uuid.UUID
	UnitOfMeasureID uuid.UUID
	SKU             string
	Barcode         string
	Name            string
	Description     string
	Cost            decimal.Decimal
	SalePrice       decimal.Decimal
	InitialStock    decimal.Decimal
	MinStock        decimal.Decimal
}

func NewProduct(clock clock.Clock, params NewProductParams) (*Product, error) {
	sku := NormalizeSKU(params.SKU)
	details := validationDetails{}
	validateSKU(details, sku)
	validateMoney(details, "cost", params.Cost)
	validateMoney(details, "sale_price", params.SalePrice)
	validateQuantity(details, "initial_stock", params.InitialStock)
	validateQuantity(details, "min_stock", params.MinStock)
	if err := details.err(); err != nil {
		return nil, err
	}

	now := clock.NowUTC()
	return &Product{
		ID:              uuid.New(),
		CompanyID:       params.CompanyID,
		CategoryID:      params.CategoryID,
		UnitOfMeasureID: params.UnitOfMeasureID,
		SKU:             sku,
		Barcode:         strings.TrimSpace(params.Barcode),
		Name:            NormalizeName(params.Name),
		Description:     strings.TrimSpace(params.Description),
		Cost:            params.Cost,
		SalePrice:       params.SalePrice,
		Stock:           params.InitialStock,
		MinStock:        params.MinStock,
		Status:          StatusActive,
		LastMovementAt:  now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

type UpdateProductParams struct {
	CategoryID      uuid.UUID
	UnitOfMeasureID uuid.UUID
	SKU             string
	Barcode         string
	Name            string
	Description     string
	Cost            decimal.Decimal
	SalePrice       decimal.Decimal
	MinStock        decimal.Decimal
}

func (p *Product) Update(clock clock.Clock, params UpdateProductParams) error {
	sku := NormalizeSKU(params.SKU)
	details := validationDetails{}
	validateSKU(details, sku)
	validateMoney(details, "cost", params.Cost)
	validateMoney(details, "sale_price", params.SalePrice)
	validateQuantity(details, "min_stock", params.MinStock)
	if err := details.err(); err != nil {
		return err
	}

	p.CategoryID = params.CategoryID
	p.UnitOfMeasureID = params.UnitOfMeasureID
	p.SKU = sku
	p.Barcode = strings.TrimSpace(params.Barcode)
	p.Name = NormalizeName(params.Name)
	p.Description = strings.TrimSpace(params.Description)
	p.Cost = params.Cost
	p.SalePrice = params.SalePrice
	p.MinStock = params.MinStock
	p.UpdatedAt = clock.NowUTC()
	return nil
}

func (p *Product) IsActive() bool {
	return p.Status == StatusActive
}

func (p *Product) Activate(clock clock.Clock) {
	p.Status = StatusActive
	p.UpdatedAt = clock.NowUTC()
}

func (p *Product) Deactivate(clock clock.Clock) {
	p.Status = StatusInactive
	p.UpdatedAt = clock.NowUTC()
}

func (p *Product) Margin() decimal.Decimal {
	if !p.SalePrice.IsPositive() {
		return decimal.Zero
	}
	return p.SalePrice.Sub(p.Cost).Div(p.SalePrice).Mul(hundred).Round(0)
}

func (p *Product) StockStatus() StockStatus {
	switch {
	case !p.Stock.IsPositive():
		return StockStatusOutOfStock
	case p.Stock.LessThanOrEqual(p.MinStock):
		return StockStatusLowStock
	default:
		return StockStatusNormal
	}
}

func NormalizeSKU(sku string) string {
	return strings.ToUpper(strings.TrimSpace(sku))
}

func validateSKU(details validationDetails, sku string) {
	if sku != "" && !skuPattern.MatchString(sku) {
		details["sku"] = "solo puede contener letras, números, '.', '-' y '_', y debe empezar por letra o número"
	}
}

func validateMoney(details validationDetails, field string, value decimal.Decimal) {
	if !value.IsPositive() {
		details[field] = "debe ser mayor que 0"
	} else if !hasAtMostDecimals(value, 2) {
		details[field] = "admite como máximo 2 decimales"
	}
}

func validateQuantity(details validationDetails, field string, value decimal.Decimal) {
	if value.IsNegative() {
		details[field] = "no puede ser negativo"
	} else if !hasAtMostDecimals(value, 3) {
		details[field] = "admite como máximo 3 decimales"
	}
}
