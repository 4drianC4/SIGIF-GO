package mapper

import (
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashsession/infrastructure/persistence/model"
)

func SessionToModel(s *entity.CashSession) *model.CashSessionModel {
	var notes *string
	if s.OpeningNotes != "" {
		n := s.OpeningNotes
		notes = &n
	}
	return &model.CashSessionModel{
		ID:             s.ID,
		CompanyID:      s.CompanyID,
		CashRegisterID: s.CashRegisterID,
		AssignmentID:   s.AssignmentID,
		UserID:         s.UserID,
		OpeningAmount:  s.OpeningAmount,
		OpeningNotes:   notes,
		Status:         string(s.Status),
		OpenedAt:       s.OpenedAt,
		ClosedAt:       s.ClosedAt,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func SessionToDomain(m *model.CashSessionModel) *entity.CashSession {
	notes := ""
	if m.OpeningNotes != nil {
		notes = *m.OpeningNotes
	}
	return &entity.CashSession{
		ID:             m.ID,
		CompanyID:      m.CompanyID,
		CashRegisterID: m.CashRegisterID,
		AssignmentID:   m.AssignmentID,
		UserID:         m.UserID,
		OpeningAmount:  m.OpeningAmount,
		OpeningNotes:   notes,
		Status:         entity.Status(m.Status),
		OpenedAt:       m.OpenedAt,
		ClosedAt:       m.ClosedAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
