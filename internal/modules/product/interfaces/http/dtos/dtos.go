package dtos

import (
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

type CreateCategoryRequest struct {
	Name         string           `json:"name" validate:"required,min=2,max=100"`
	ParentID     *string          `json:"parent_id" validate:"omitempty,uuid"`
	Description  string           `json:"description" validate:"omitempty,max=500"`
	DefaultTax   *string          `json:"default_tax" validate:"omitempty,max=60"`
	TargetMargin *decimal.Decimal `json:"target_margin"`
}

func (r *CreateCategoryRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
	r.ParentID = trimOptional(r.ParentID)
	r.DefaultTax = trimOptional(r.DefaultTax)
}

func (r *CreateCategoryRequest) ParentUUID() *uuid.UUID {
	return parseOptionalUUID(r.ParentID)
}

type CreateProductRequest struct {
	Name            string           `json:"name" validate:"required,min=2,max=150"`
	SKU             string           `json:"sku" validate:"omitempty,max=50"`
	Barcode         string           `json:"barcode" validate:"omitempty,numeric,min=8,max=14"`
	Description     string           `json:"description" validate:"omitempty,max=1000"`
	CategoryID      string           `json:"category_id" validate:"required,uuid"`
	UnitOfMeasureID string           `json:"unit_of_measure_id" validate:"required,uuid"`
	Cost            *decimal.Decimal `json:"cost" validate:"required"`
	SalePrice       *decimal.Decimal `json:"sale_price" validate:"required"`
	InitialStock    *decimal.Decimal `json:"initial_stock"`
	MinStock        *decimal.Decimal `json:"min_stock"`
}

func (r *CreateProductRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.SKU = strings.TrimSpace(r.SKU)
	r.Barcode = strings.TrimSpace(r.Barcode)
	r.Description = strings.TrimSpace(r.Description)
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.UnitOfMeasureID = strings.TrimSpace(r.UnitOfMeasureID)
}

type ListProductsQuery struct {
	Search     string `query:"search" json:"search" validate:"omitempty,max=100"`
	CategoryID string `query:"category_id" json:"category_id" validate:"omitempty,uuid"`
	Status     string `query:"status" json:"status" validate:"omitempty,oneof=active inactive"`
}

func (q *ListProductsQuery) CategoryUUID() *uuid.UUID {
	return parseOptionalUUID(trimOptional(&q.CategoryID))
}

func (q *ListProductsQuery) StatusFilter() *entity.Status {
	if q.Status == "" {
		return nil
	}
	status := entity.Status(q.Status)
	return &status
}

type ValidateDuplicateQuery struct {
	Name      string `query:"name" json:"name" validate:"omitempty,max=150"`
	SKU       string `query:"sku" json:"sku" validate:"omitempty,max=50"`
	Barcode   string `query:"barcode" json:"barcode" validate:"omitempty,max=14"`
	ExcludeID string `query:"exclude_id" json:"exclude_id" validate:"omitempty,uuid"`
}

func (q *ValidateDuplicateQuery) ExcludeUUID() uuid.UUID {
	if id := parseOptionalUUID(trimOptional(&q.ExcludeID)); id != nil {
		return *id
	}
	return uuid.Nil
}

func OrZero(d *decimal.Decimal) decimal.Decimal {
	if d == nil {
		return decimal.Zero
	}
	return *d
}

func trimOptional(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func parseOptionalUUID(s *string) *uuid.UUID {
	if s == nil {
		return nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	return &id
}

type UpdateProductRequest struct {
	Name            string           `json:"name" validate:"required,min=2,max=150"`
	SKU             string           `json:"sku" validate:"omitempty,max=50"`
	Barcode         string           `json:"barcode" validate:"omitempty,numeric,min=8,max=14"`
	Description     string           `json:"description" validate:"omitempty,max=1000"`
	CategoryID      string           `json:"category_id" validate:"required,uuid"`
	UnitOfMeasureID string           `json:"unit_of_measure_id" validate:"required,uuid"`
	Cost            *decimal.Decimal `json:"cost" validate:"required"`
	SalePrice       *decimal.Decimal `json:"sale_price" validate:"required"`
	MinStock        *decimal.Decimal `json:"min_stock"`
}

func (r *UpdateProductRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.SKU = strings.TrimSpace(r.SKU)
	r.Barcode = strings.TrimSpace(r.Barcode)
	r.Description = strings.TrimSpace(r.Description)
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.UnitOfMeasureID = strings.TrimSpace(r.UnitOfMeasureID)
}
