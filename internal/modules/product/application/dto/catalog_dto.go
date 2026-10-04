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
