package dto

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
)

const timeLayout = "2006-01-02T15:04:05Z07:00"

type Category struct {
	ID          string        `json:"id"`
	CompanyID   string        `json:"company_id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Status      entity.Status `json:"status"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
}

type Product struct {
	ID            string               `json:"id"`
	CompanyID     string               `json:"company_id"`
	CategoryID    string               `json:"category_id"`
	SKU           string               `json:"sku"`
	Barcode       string               `json:"barcode,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	UnitOfMeasure entity.UnitOfMeasure `json:"unit_of_measure"`
	CostPrice     string               `json:"cost_price"`
	SalePrice     string               `json:"sale_price"`
	Status        entity.Status        `json:"status"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
}

func FromCategory(c *entity.Category) Category {
	return Category{
		ID:          c.ID.String(),
		CompanyID:   c.CompanyID.String(),
		Name:        c.Name,
		Description: c.Description,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt.Format(timeLayout),
		UpdatedAt:   c.UpdatedAt.Format(timeLayout),
	}
}

func FromProduct(p *entity.Product) Product {
	return Product{
		ID:            p.ID.String(),
		CompanyID:     p.CompanyID.String(),
		CategoryID:    p.CategoryID.String(),
		SKU:           p.SKU,
		Barcode:       p.Barcode,
		Name:          p.Name,
		Description:   p.Description,
		UnitOfMeasure: p.UnitOfMeasure,
		CostPrice:     p.CostPrice.StringFixed(2),
		SalePrice:     p.SalePrice.StringFixed(2),
		Status:        p.Status,
		CreatedAt:     p.CreatedAt.Format(timeLayout),
		UpdatedAt:     p.UpdatedAt.Format(timeLayout),
	}
}

func FromProductList(products []*entity.Product) []Product {
	result := make([]Product, 0, len(products))
	for _, p := range products {
		result = append(result, FromProduct(p))
	}
	return result
}

type ProductDetail struct {
	Product
	StockTotal string `json:"stock_total"`
}

func FromProductDetail(p *entity.ProductDetail) ProductDetail {
	return ProductDetail{Product: FromProduct(p.Product), StockTotal: p.StockTotal.StringFixed(3)}
}

type PriceHistoryEntry struct {
	ID            string `json:"id"`
	UserFullName  string `json:"user_full_name"`
	UserRole      string `json:"user_role"`
	PreviousCost  string `json:"previous_cost"`
	NewCost       string `json:"new_cost"`
	PreviousPrice string `json:"previous_price"`
	NewPrice      string `json:"new_price"`
	MarginPct     string `json:"margin_pct"`
	Reason        string `json:"reason,omitempty"`
	CreatedAt     string `json:"created_at"`
}

func FromPriceHistory(entries []*entity.PriceHistoryEntry) []PriceHistoryEntry {
	result := make([]PriceHistoryEntry, 0, len(entries))
	for _, e := range entries {
		result = append(result, PriceHistoryEntry{
			ID:            e.ID.String(),
			UserFullName:  e.UserFullName,
			UserRole:      e.UserRole,
			PreviousCost:  e.PreviousCost.StringFixed(2),
			NewCost:       e.NewCost.StringFixed(2),
			PreviousPrice: e.PreviousPrice.StringFixed(2),
			NewPrice:      e.NewPrice.StringFixed(2),
			MarginPct:     e.MarginPct.String() + "%",
			Reason:        e.Reason,
			CreatedAt:     e.CreatedAt.Format(timeLayout),
		})
	}
	return result
}
