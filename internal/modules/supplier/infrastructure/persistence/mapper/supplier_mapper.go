package mapper

import (
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/model"
)

func ToModel(s *entity.Supplier) *model.SupplierModel {
	return &model.SupplierModel{ID: s.ID, CompanyID: s.CompanyID, TradeName: s.TradeName,
		LegalName: s.LegalName, TaxID: s.TaxID, Contact: s.Contact, Phone: s.Phone, Email: s.Email,
		Address: s.Address, MainIndustry: s.MainIndustry, PaymentTerms: s.PaymentTerms,
		LeadTimeDays: s.LeadTimeDays, MinimumOrder: s.MinimumOrder,
		Status: string(s.Status), CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
}

func ToDomain(m *model.SupplierModel) *entity.Supplier {
	return &entity.Supplier{ID: m.ID, CompanyID: m.CompanyID,
		Fields: entity.Fields{TradeName: m.TradeName, LegalName: m.LegalName, TaxID: m.TaxID,
			Contact: m.Contact, Phone: m.Phone, Email: m.Email, Address: m.Address,
			MainIndustry: m.MainIndustry, PaymentTerms: m.PaymentTerms,
			LeadTimeDays: m.LeadTimeDays, MinimumOrder: m.MinimumOrder},
		Status: entity.Status(m.Status), CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

func AssociationToDomain(a *model.ProductSupplierModel) *entity.ProductSupplier {
	return &entity.ProductSupplier{ID: a.ID, CompanyID: a.CompanyID, SupplierID: a.SupplierID,
		ProductID: a.ProductID, LeadTimeDays: a.LeadTimeDays, CurrentCost: a.CurrentCost,
		SupplierCode: a.SupplierCode, Preferred: a.Preferred,
		CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt}
}

func AssociationToModel(a *entity.ProductSupplier) *model.ProductSupplierModel {
	return &model.ProductSupplierModel{ID: a.ID, CompanyID: a.CompanyID, SupplierID: a.SupplierID,
		ProductID: a.ProductID, LeadTimeDays: a.LeadTimeDays, CurrentCost: a.CurrentCost,
		SupplierCode: a.SupplierCode, Preferred: a.Preferred, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt}
}
