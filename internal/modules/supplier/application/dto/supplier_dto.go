package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
)

type Money decimal.Decimal

func (m Money) MarshalJSON() ([]byte, error) { return []byte(decimal.Decimal(m).StringFixed(2)), nil }

type Supplier struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	TradeName    *string       `json:"trade_name"`
	LegalName    string        `json:"legal_name"`
	TaxID        *string       `json:"tax_id"`
	Contact      *string       `json:"contact"`
	Phone        *string       `json:"phone"`
	Email        *string       `json:"email"`
	Address      *string       `json:"address"`
	MainIndustry *string       `json:"main_industry"`
	Status       entity.Status `json:"status"`
	LeadTimeDays *int          `json:"lead_time_days"`
	PaymentTerms *string       `json:"payment_terms"`
	MinimumOrder *Money        `json:"minimum_order"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    *string       `json:"updated_at"`
}

type SupplierListItem struct {
	Supplier
	ProductsCount int64 `json:"products_count"`
}
type Summary struct {
	ActiveSuppliers int64 `json:"active_suppliers"`
}

type Association struct {
	ID           string  `json:"id"`
	SupplierID   string  `json:"supplier_id"`
	ProductID    string  `json:"product_id"`
	SupplierCode *string `json:"supplier_code"`
	LeadTimeDays *int    `json:"lead_time_days"`
	CurrentCost  *Money  `json:"current_cost"`
	Preferred    bool    `json:"preferred"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    *string `json:"updated_at"`
}
type ProductItem struct {
	Association
	Name                  string  `json:"name"`
	SKU                   *string `json:"sku"`
	EffectiveLeadTimeDays *int    `json:"effective_lead_time_days"`
}
type FormOptions struct {
	Industries   []string `json:"industries"`
	PaymentTerms []string `json:"payment_terms"`
}

func timestamp(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
func money(d *decimal.Decimal) *Money {
	if d == nil {
		return nil
	}
	m := Money(*d)
	return &m
}

func FromSupplier(s *entity.Supplier) Supplier {
	return Supplier{ID: s.ID.String(), Name: s.Name(), TradeName: s.TradeName, LegalName: s.LegalName,
		TaxID: s.TaxID, Contact: s.Contact, Phone: s.Phone, Email: s.Email, Address: s.Address,
		MainIndustry: s.MainIndustry, Status: s.Status, LeadTimeDays: s.LeadTimeDays,
		PaymentTerms: s.PaymentTerms, MinimumOrder: money(s.MinimumOrder),
		CreatedAt: s.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: timestamp(s.UpdatedAt)}
}
func FromList(items []repository.SupplierListItem) []SupplierListItem {
	result := make([]SupplierListItem, len(items))
	for i, item := range items {
		result[i] = SupplierListItem{Supplier: FromSupplier(item.Supplier), ProductsCount: item.ProductsCount}
	}
	return result
}
func FromSummary(s repository.SupplierSummary) Summary {
	return Summary{ActiveSuppliers: s.ActiveSuppliers}
}
func FromAssociation(a *entity.ProductSupplier) Association {
	return Association{ID: a.ID.String(), SupplierID: a.SupplierID.String(), ProductID: a.ProductID.String(),
		LeadTimeDays: a.LeadTimeDays, CurrentCost: money(a.CurrentCost), SupplierCode: a.SupplierCode, Preferred: a.Preferred,
		CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: timestamp(a.UpdatedAt)}
}
func FromProducts(items []repository.ProductItem) []ProductItem {
	result := make([]ProductItem, len(items))
	for i, item := range items {
		result[i] = ProductItem{Association: FromAssociation(item.Association), Name: item.Name, SKU: item.SKU, EffectiveLeadTimeDays: item.EffectiveLeadTimeDays}
	}
	return result
}

type SupplierReference struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	LegalName    string        `json:"legal_name"`
	TaxID        *string       `json:"tax_id"`
	MainIndustry *string       `json:"main_industry"`
	Status       entity.Status `json:"status"`
	LeadTimeDays *int          `json:"lead_time_days"`
}
type SupplierItem struct {
	Association
	Supplier              SupplierReference `json:"supplier"`
	EffectiveLeadTimeDays *int              `json:"effective_lead_time_days"`
}
type SupplierOption struct {
	SupplierReference
	IsAssociated bool `json:"is_associated"`
	Selectable   bool `json:"selectable"`
}
type Quantity decimal.Decimal

func (q Quantity) MarshalJSON() ([]byte, error) { return []byte(decimal.Decimal(q).String()), nil }

type CategoryReference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}
type ProductOption struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	SKU              *string           `json:"sku"`
	Barcode          *string           `json:"barcode"`
	Status           string            `json:"status"`
	Category         CategoryReference `json:"category"`
	Stock            Quantity          `json:"stock"`
	UnitAbbreviation string            `json:"unit_abbreviation"`
	IsAssociated     bool              `json:"is_associated"`
	Selectable       bool              `json:"selectable"`
}

func supplierRef(s *entity.Supplier) SupplierReference {
	return SupplierReference{ID: s.ID.String(), Name: s.Name(), LegalName: s.LegalName, TaxID: s.TaxID, MainIndustry: s.MainIndustry, Status: s.Status, LeadTimeDays: s.LeadTimeDays}
}
func FromProductSuppliers(items []repository.SupplierItem) []SupplierItem {
	result := make([]SupplierItem, len(items))
	for i, v := range items {
		result[i] = SupplierItem{Association: FromAssociation(v.Association), Supplier: supplierRef(v.Supplier), EffectiveLeadTimeDays: v.EffectiveLeadTimeDays}
	}
	return result
}
func FromSupplierOptions(items []repository.SupplierOption) []SupplierOption {
	result := make([]SupplierOption, len(items))
	for i, v := range items {
		result[i] = SupplierOption{SupplierReference: supplierRef(v.Supplier), IsAssociated: v.IsAssociated, Selectable: v.Selectable}
	}
	return result
}
func FromProductOptions(items []repository.ProductOption) []ProductOption {
	result := make([]ProductOption, len(items))
	for i, v := range items {
		result[i] = ProductOption{ID: v.ID.String(), Name: v.Name, SKU: v.SKU, Barcode: v.Barcode, Status: v.Status,
			Category: CategoryReference{ID: v.CategoryID.String(), Name: v.CategoryName, Path: v.CategoryPath}, Stock: Quantity(v.Stock), UnitAbbreviation: v.UnitAbbreviation, IsAssociated: v.IsAssociated, Selectable: v.Selectable}
	}
	return result
}
func FromAssociations(items []*entity.ProductSupplier) []Association {
	result := make([]Association, len(items))
	for i, v := range items {
		result[i] = FromAssociation(v)
	}
	return result
}
