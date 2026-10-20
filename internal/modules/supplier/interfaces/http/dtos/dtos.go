package dtos

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
	"strings"
)

type SupplierRequest struct {
	TradeName    *string          `json:"trade_name"`
	LegalName    string           `json:"legal_name" validate:"required,max=160"`
	TaxID        *string          `json:"tax_id"`
	Contact      *string          `json:"contact"`
	Phone        *string          `json:"phone"`
	Email        *string          `json:"email"`
	Address      *string          `json:"address"`
	MainIndustry *string          `json:"main_industry"`
	PaymentTerms *string          `json:"payment_terms"`
	LeadTimeDays *int             `json:"lead_time_days"`
	MinimumOrder *decimal.Decimal `json:"minimum_order"`
}

func (r *SupplierRequest) Normalize() { r.LegalName = strings.TrimSpace(r.LegalName) }

func (r SupplierRequest) Fields() entity.Fields {
	return entity.Fields{TradeName: r.TradeName, LegalName: r.LegalName, TaxID: r.TaxID,
		Contact: r.Contact, Phone: r.Phone, Email: r.Email, Address: r.Address,
		MainIndustry: r.MainIndustry, PaymentTerms: r.PaymentTerms, LeadTimeDays: r.LeadTimeDays, MinimumOrder: r.MinimumOrder}
}

type ListQuery struct {
	Search       string `query:"search" json:"search" validate:"omitempty,max=160"`
	Status       string `query:"status" json:"status" validate:"omitempty,oneof=active inactive"`
	MainIndustry string `query:"main_industry" json:"main_industry" validate:"omitempty,max=80"`
	SortBy       string `query:"sort_by" json:"sort_by" validate:"omitempty,oneof=name legal_name created_at lead_time_days"`
	SortOrder    string `query:"sort_order" json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

func (q ListQuery) Filter() repository.ListFilter {
	f := repository.ListFilter{Search: q.Search, MainIndustry: q.MainIndustry, SortBy: q.SortBy, SortOrder: q.SortOrder}
	if q.Status != "" {
		s := entity.Status(q.Status)
		f.Status = &s
	}
	return f
}

type ProductsQuery struct {
	Search string `query:"search" json:"search" validate:"omitempty,max=160"`
}

type ProductOptionsQuery struct {
	Search     string `query:"search" json:"search" validate:"omitempty,max=160"`
	CategoryID string `query:"category_id" json:"category_id" validate:"omitempty,uuid"`
}

func (q ProductOptionsQuery) Filter() repository.ProductOptionFilter {
	f := repository.ProductOptionFilter{Search: q.Search}
	if q.CategoryID != "" {
		id, _ := uuid.Parse(q.CategoryID)
		f.CategoryID = &id
	}
	return f
}

type AssociateProductRequest struct {
	ProductID    string           `json:"product_id" validate:"required,uuid"`
	SupplierCode *string          `json:"supplier_code"`
	LeadTimeDays *int             `json:"lead_time_days"`
	CurrentCost  *decimal.Decimal `json:"current_cost" validate:"required"`
	Preferred    *bool            `json:"preferred"`
}
type AssociateProductsRequest struct {
	ProductIDs []string `json:"product_ids" validate:"required,min=1,max=100,dive,uuid"`
}
type UpdateAssociationRequest struct {
	SupplierCode *string          `json:"supplier_code"`
	LeadTimeDays *int             `json:"lead_time_days"`
	CurrentCost  *decimal.Decimal `json:"current_cost"`
	Preferred    *bool            `json:"preferred"`
}
