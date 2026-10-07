package dtos

import (
	"strings"

	"github.com/shopspring/decimal"
)

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

func (r *CreateCategoryRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
}

type CreateProductRequest struct {
	CategoryID    string           `json:"category_id" validate:"required,uuid"`
	SKU           string           `json:"sku" validate:"required,max=50"`
	Barcode       string           `json:"barcode" validate:"omitempty,numeric,min=8,max=14"`
	Name          string           `json:"name" validate:"required,min=2,max=150"`
	Description   string           `json:"description" validate:"omitempty,max=1000"`
	UnitOfMeasure string           `json:"unit_of_measure" validate:"required,oneof=unit kg g l ml m box pack"`
	CostPrice     *decimal.Decimal `json:"cost_price"`
	SalePrice     *decimal.Decimal `json:"sale_price" validate:"required"`
}

func (r *CreateProductRequest) Normalize() {
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.SKU = strings.TrimSpace(r.SKU)
	r.Barcode = strings.TrimSpace(r.Barcode)
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
	r.UnitOfMeasure = strings.TrimSpace(r.UnitOfMeasure)
}

type UpdateProductRequest struct {
	CategoryID    string           `json:"category_id" validate:"required,uuid"`
	SKU           string           `json:"sku" validate:"required,max=50"`
	Barcode       string           `json:"barcode" validate:"omitempty,numeric,min=8,max=14"`
	Name          string           `json:"name" validate:"required,min=2,max=150"`
	Description   string           `json:"description" validate:"omitempty,max=1000"`
	UnitOfMeasure string           `json:"unit_of_measure" validate:"required,oneof=unit kg g l ml m box pack"`
	CostPrice     *decimal.Decimal `json:"cost_price"`
	SalePrice     *decimal.Decimal `json:"sale_price" validate:"required"`
}

func (r *UpdateProductRequest) Normalize() {
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.SKU = strings.TrimSpace(r.SKU)
	r.Barcode = strings.TrimSpace(r.Barcode)
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
	r.UnitOfMeasure = strings.TrimSpace(r.UnitOfMeasure)
}
