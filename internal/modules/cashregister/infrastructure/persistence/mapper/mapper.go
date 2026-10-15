package mapper

import (
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/model"
)

func RegisterToDomain(m *model.CashRegisterModel) *entity.CashRegister {
	return &entity.CashRegister{
		ID:        m.ID,
		CompanyID: m.CompanyID,
		Code:      m.Code,
		Name:      m.Name,
		Status:    entity.CashRegisterStatus(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func AssignmentToModel(a *entity.Assignment) *model.CashRegisterAssignmentModel {
	return &model.CashRegisterAssignmentModel{
		ID:             a.ID,
		CompanyID:      a.CompanyID,
		CashRegisterID: a.CashRegisterID,
		UserID:         a.UserID,
		Status:         string(a.Status),
		AssignedAt:     a.AssignedAt,
		ReleasedAt:     a.ReleasedAt,
		ReleasedBy:     a.ReleasedBy,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

func AssignmentToDomain(m *model.CashRegisterAssignmentModel) *entity.Assignment {
	return &entity.Assignment{
		ID:             m.ID,
		CompanyID:      m.CompanyID,
		CashRegisterID: m.CashRegisterID,
		UserID:         m.UserID,
		Status:         entity.AssignmentStatus(m.Status),
		AssignedAt:     m.AssignedAt,
		ReleasedAt:     m.ReleasedAt,
		ReleasedBy:     m.ReleasedBy,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
