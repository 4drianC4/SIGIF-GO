package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

type Money decimal.Decimal

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(decimal.Decimal(m).StringFixed(2)), nil
}

type Quantity decimal.Decimal

func (q Quantity) MarshalJSON() ([]byte, error) {
	return []byte(decimal.Decimal(q).String()), nil
}

type CategoryRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	SKU             *string            `json:"sku"`
	Barcode         *string            `json:"barcode"`
	Description     string             `json:"description,omitempty"`
	Category        CategoryRef        `json:"category"`
	UnitOfMeasureID string             `json:"unit_of_measure_id"`
	Cost            Money              `json:"cost"`
	Price           Money              `json:"price"`
	Margin          int64              `json:"margin"`
	Stock           Quantity           `json:"stock"`
	MinStock        Quantity           `json:"min_stock"`
	StockStatus     entity.StockStatus `json:"stock_status"`
	Status          entity.Status      `json:"status"`
	CreatedAt       string             `json:"created_at"`
	UpdatedAt       string             `json:"updated_at"`
}

type Category struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	ParentID      *string       `json:"parent_id"`
	Description   string        `json:"description,omitempty"`
	ProductsCount int64         `json:"products_count"`
	DefaultTax    *string       `json:"default_tax"`
	TargetMargin  *Quantity     `json:"target_margin"`
	Status        entity.Status `json:"status"`
}

type ProductSummary struct {
	ActiveProducts   int64 `json:"active_products"`
	InventoryValue   Money `json:"inventory_value"`
	LowStock         int64 `json:"low_stock"`
	NoMovement90Days int64 `json:"no_movement_90_days"`
}

type DuplicateCheck struct {
	Exists bool    `json:"exists"`
	Field  *string `json:"field"`
}

type UnitOfMeasure struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

type Tax struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Percentage Quantity `json:"percentage"`
}

func FromProduct(item repository.ProductListItem) Product {
	p := item.Product
	return Product{
		ID:              p.ID.String(),
		Name:            p.Name,
		SKU:             optional(p.SKU),
		Barcode:         optional(p.Barcode),
		Description:     p.Description,
		Category:        CategoryRef{ID: p.CategoryID.String(), Name: item.CategoryName},
		UnitOfMeasureID: p.UnitOfMeasureID.String(),
		Cost:            Money(p.Cost),
		Price:           Money(p.SalePrice),
		Margin:          p.Margin().IntPart(),
		Stock:           Quantity(p.Stock),
		MinStock:        Quantity(p.MinStock),
		StockStatus:     p.StockStatus(),
		Status:          p.Status,
		CreatedAt:       p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func FromProductList(items []repository.ProductListItem) []Product {
	result := make([]Product, len(items))
	for i, item := range items {
		result[i] = FromProduct(item)
	}
	return result
}

func FromCategory(c *entity.Category, productsCount int64, defaultTaxName *string) Category {
	var parentID *string
	if c.ParentID != nil {
		s := c.ParentID.String()
		parentID = &s
	}
	var targetMargin *Quantity
	if c.TargetMargin != nil {
		m := Quantity(*c.TargetMargin)
		targetMargin = &m
	}
	return Category{
		ID:            c.ID.String(),
		Name:          c.Name,
		ParentID:      parentID,
		Description:   c.Description,
		ProductsCount: productsCount,
		DefaultTax:    defaultTaxName,
		TargetMargin:  targetMargin,
		Status:        c.Status,
	}
}

func FromCreatedCategory(created *service.CreatedCategory) Category {
	return FromCategory(created.Category, 0, created.DefaultTaxName)
}

func FromCategoryList(items []repository.CategoryListItem) []Category {
	result := make([]Category, len(items))
	for i, item := range items {
		result[i] = FromCategory(item.Category, item.ProductsCount, item.DefaultTaxName)
	}
	return result
}

func FromSummary(s repository.ProductSummary) ProductSummary {
	return ProductSummary{
		ActiveProducts:   s.ActiveProducts,
		InventoryValue:   Money(s.InventoryValue),
		LowStock:         s.LowStock,
		NoMovement90Days: s.WithoutMovement,
	}
}

func FromDuplicateResult(r *service.DuplicateResult) DuplicateCheck {
	return DuplicateCheck{Exists: r.Exists, Field: optional(r.Field)}
}

func FromUnits(units []*entity.UnitOfMeasure) []UnitOfMeasure {
	result := make([]UnitOfMeasure, len(units))
	for i, u := range units {
		result[i] = UnitOfMeasure{ID: u.ID.String(), Name: u.Name, Abbreviation: u.Abbreviation}
	}
	return result
}

func FromTaxes(taxes []*entity.Tax) []Tax {
	result := make([]Tax, len(taxes))
	for i, t := range taxes {
		result[i] = Tax{ID: t.ID.String(), Name: t.Name, Percentage: Quantity(t.Percentage)}
	}
	return result
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
