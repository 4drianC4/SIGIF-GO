package mapper

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/inventory/infrastructure/persistence/model"
)

func MovementToModel(m *entity.Movement) *model.MovementModel {
	return &model.MovementModel{
		ID:         m.ID,
		CompanyID:  m.CompanyID,
		ProductID:  m.ProductID,
		BranchID:   m.BranchID,
		LocationID: m.LocationID,
		Type:       string(m.Type),
		Quantity:   m.Quantity,
		Reason:     nullable(m.Reason),
		UserID:     nullableUUID(m.UserID),
		CreatedAt:  m.CreatedAt,
	}
}

func MovementToDomain(m *model.MovementModel) *entity.Movement {
	return &entity.Movement{
		ID:         m.ID,
		CompanyID:  m.CompanyID,
		ProductID:  m.ProductID,
		BranchID:   m.BranchID,
		LocationID: m.LocationID,
		Type:       entity.MovementType(m.Type),
		Quantity:   m.Quantity,
		Reason:     deref(m.Reason),
		UserID:     derefUUID(m.UserID),
		CreatedAt:  m.CreatedAt,
	}
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

func nullableUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func derefUUID(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
}
