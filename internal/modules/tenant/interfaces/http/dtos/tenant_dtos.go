package dtos

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
)

type TenantResponse struct {
	ID            uuid.UUID             `json:"id"`
	Name          string                `json:"name"`
	Slug          string                `json:"slug"`
	BusinessType  entity.BusinessType   `json:"business_type"`
	IsActive      bool                  `json:"is_active"`
	Settings      entity.TenantSettings `json:"settings"`
	CreatedAt     string                `json:"created_at"`
	UpdatedAt     string                `json:"updated_at"`
}

func ToTenantResponse(t *entity.Tenant) TenantResponse {
	return TenantResponse{
		ID:           t.ID,
		Name:         t.Name,
		Slug:         t.Slug,
		BusinessType: t.BusinessType,
		IsActive:     t.IsActive,
		Settings:     t.Settings,
		CreatedAt:    t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToTenantResponseList(tenants []*entity.Tenant) []TenantResponse {
	result := make([]TenantResponse, len(tenants))
	for i, t := range tenants {
		result[i] = ToTenantResponse(t)
	}
	return result
}

type CreateTenantRequest struct {
	Name         string `json:"name" validate:"required,min=2,max=255"`
	Slug         string `json:"slug" validate:"required,min=2,max=100,alphanum"`
	BusinessType string `json:"business_type" validate:"required,oneof=minimarket hardware pharmacy"`
}

type UpdateTenantRequest struct {
	Name     string                 `json:"name" validate:"required,min=2,max=255"`
	Settings entity.TenantSettings  `json:"settings"`
}

type TenantListResponse struct {
	Tenants []TenantResponse `json:"tenants"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}