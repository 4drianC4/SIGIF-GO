package mapper

import (
	"time"

	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
)

func CategoryToModel(c *entity.Category) *model.CategoryModel {
	return &model.CategoryModel{
		ID:          c.ID,
		TenantID:    c.TenantID,
		Name:        c.Name,
		Description: c.Description,
		Status:      string(c.Status),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		DeletedAt:   toDeletedAt(c.DeletedAt),
	}
}

func CategoryToDomain(m *model.CategoryModel) *entity.Category {
	return &entity.Category{
		ID:          m.ID,
		TenantID:    m.TenantID,
		Name:        m.Name,
		Description: m.Description,
		Status:      entity.Status(m.Status),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   fromDeletedAt(m.DeletedAt),
	}
}

// ProductToModel guarda el código de barras vacío como NULL para que no choque con el índice único.
func ProductToModel(p *entity.Product) *model.ProductModel {
	var barcode *string
	if p.Barcode != "" {
		b := p.Barcode
		barcode = &b
	}

	return &model.ProductModel{
		ID:            p.ID,
		TenantID:      p.TenantID,
		CategoryID:    p.CategoryID,
		SKU:           p.SKU,
		Barcode:       barcode,
		Name:          p.Name,
		Description:   p.Description,
		UnitOfMeasure: string(p.UnitOfMeasure),
		CostPrice:     p.CostPrice,
		SalePrice:     p.SalePrice,
		Status:        string(p.Status),
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		DeletedAt:     toDeletedAt(p.DeletedAt),
	}
}

func ProductToDomain(m *model.ProductModel) *entity.Product {
	barcode := ""
	if m.Barcode != nil {
		barcode = *m.Barcode
	}

	return &entity.Product{
		ID:            m.ID,
		TenantID:      m.TenantID,
		CategoryID:    m.CategoryID,
		SKU:           m.SKU,
		Barcode:       barcode,
		Name:          m.Name,
		Description:   m.Description,
		UnitOfMeasure: entity.UnitOfMeasure(m.UnitOfMeasure),
		CostPrice:     m.CostPrice,
		SalePrice:     m.SalePrice,
		Status:        entity.Status(m.Status),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		DeletedAt:     fromDeletedAt(m.DeletedAt),
	}
}

func toDeletedAt(t *time.Time) gorm.DeletedAt {
	if t == nil {
		return gorm.DeletedAt{}
	}
	return gorm.DeletedAt{Time: *t, Valid: true}
}

func fromDeletedAt(d gorm.DeletedAt) *time.Time {
	if !d.Valid {
		return nil
	}
	t := d.Time
	return &t
}
