package event

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/events"
)

type TenantCreatedEvent struct {
	*events.BaseEvent
	Tenant *entity.Tenant
}

func NewTenantCreatedEvent(tenant *entity.Tenant) *TenantCreatedEvent {
	return &TenantCreatedEvent{
		BaseEvent: events.NewEvent("tenant.created", tenant),
		Tenant:    tenant,
	}
}

type TenantUpdatedEvent struct {
	*events.BaseEvent
	Tenant *entity.Tenant
}

func NewTenantUpdatedEvent(tenant *entity.Tenant) *TenantUpdatedEvent {
	return &TenantUpdatedEvent{
		BaseEvent: events.NewEvent("tenant.updated", tenant),
		Tenant:    tenant,
	}
}

type TenantDeletedEvent struct {
	*events.BaseEvent
	TenantID uuid.UUID
}

func NewTenantDeletedEvent(tenantID uuid.UUID) *TenantDeletedEvent {
	return &TenantDeletedEvent{
		BaseEvent: events.NewEvent("tenant.deleted", map[string]any{"tenant_id": tenantID}),
		TenantID:  tenantID,
	}
}

type TenantActivatedEvent struct {
	*events.BaseEvent
	Tenant *entity.Tenant
}

func NewTenantActivatedEvent(tenant *entity.Tenant) *TenantActivatedEvent {
	return &TenantActivatedEvent{
		BaseEvent: events.NewEvent("tenant.activated", tenant),
		Tenant:    tenant,
	}
}

type TenantDeactivatedEvent struct {
	*events.BaseEvent
	Tenant *entity.Tenant
}

func NewTenantDeactivatedEvent(tenant *entity.Tenant) *TenantDeactivatedEvent {
	return &TenantDeactivatedEvent{
		BaseEvent: events.NewEvent("tenant.deactivated", tenant),
		Tenant:    tenant,
	}
}