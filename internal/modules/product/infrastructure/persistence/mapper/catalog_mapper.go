package mapper

import (
	"time"

	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
)

func CategoryToModel(c *entity.Category) *model.CategoryModel {
	return &model.CategoryModel{
		ID:           c.ID,
		CompanyID:    c.CompanyID,
		ParentID:     c.ParentID,
		Name:         c.Name,
		Description:  c.Description,
		DefaultTaxID: c.DefaultTaxID,
		TargetMargin: c.TargetMargin,
		Status:       string(c.Status),
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		DeletedAt:    toDeletedAt(c.DeletedAt),
	}
}

func CategoryToDomain(m *model.CategoryModel) *entity.Category {
	return &entity.Category{
		ID:           m.ID,
		CompanyID:    m.CompanyID,
		ParentID:     m.ParentID,
		Name:         m.Name,
		Description:  m.Description,
		DefaultTaxID: m.DefaultTaxID,
		TargetMargin: m.TargetMargin,
		Status:       entity.Status(m.Status),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		DeletedAt:    fromDeletedAt(m.DeletedAt),
	}
}

func ProductToModel(p *entity.Product) *model.ProductModel {
	return &model.ProductModel{
		ID:              p.ID,
		CompanyID:       p.CompanyID,
		CategoryID:      p.CategoryID,
		UnitOfMeasureID: p.UnitOfMeasureID,
		SKU:             nullable(p.SKU),
		Barcode:         nullable(p.Barcode),
		Name:            p.Name,
		Description:     p.Description,
		Cost:            p.Cost,
		SalePrice:       p.SalePrice,
		Stock:           p.Stock,
		MinStock:        p.MinStock,
		Status:          string(p.Status),
		LastMovementAt:  p.LastMovementAt,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		DeletedAt:       toDeletedAt(p.DeletedAt),
	}
}

func ProductToDomain(m *model.ProductModel) *entity.Product {
	return &entity.Product{
		ID:              m.ID,
		CompanyID:       m.CompanyID,
		CategoryID:      m.CategoryID,
		UnitOfMeasureID: m.UnitOfMeasureID,
		SKU:             deref(m.SKU),
		Barcode:         deref(m.Barcode),
		Name:            m.Name,
		Description:     m.Description,
		Cost:            m.Cost,
		SalePrice:       m.SalePrice,
		Stock:           m.Stock,
		MinStock:        m.MinStock,
		Status:          entity.Status(m.Status),
		LastMovementAt:  m.LastMovementAt,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		DeletedAt:       fromDeletedAt(m.DeletedAt),
	}
}

func UnitToDomain(m *model.UnitOfMeasureModel) *entity.UnitOfMeasure {
	return &entity.UnitOfMeasure{ID: m.ID, Name: m.Name, Abbreviation: m.Abbreviation}
}

func TaxToDomain(m *model.TaxModel) *entity.Tax {
	return &entity.Tax{ID: m.ID, Name: m.Name, Percentage: m.Percentage}
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
