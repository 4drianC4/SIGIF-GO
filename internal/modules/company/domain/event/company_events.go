package event

import (
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/events"
)

type CompanyCreatedEvent struct {
	*events.BaseEvent
	Company *entity.Company
}

func NewCompanyCreatedEvent(company *entity.Company) *CompanyCreatedEvent {
	return &CompanyCreatedEvent{
		BaseEvent: events.NewEvent("company.created", company),
		Company:   company,
	}
}

type CompanyUpdatedEvent struct {
	*events.BaseEvent
	Company *entity.Company
}

func NewCompanyUpdatedEvent(company *entity.Company) *CompanyUpdatedEvent {
	return &CompanyUpdatedEvent{
		BaseEvent: events.NewEvent("company.updated", company),
		Company:   company,
	}
}

type CompanyDeletedEvent struct {
	*events.BaseEvent
	CompanyID uuid.UUID
}

func NewCompanyDeletedEvent(companyID uuid.UUID) *CompanyDeletedEvent {
	return &CompanyDeletedEvent{
		BaseEvent: events.NewEvent("company.deleted", map[string]any{"company_id": companyID}),
		CompanyID: companyID,
	}
}

type CompanyActivatedEvent struct {
	*events.BaseEvent
	Company *entity.Company
}

func NewCompanyActivatedEvent(company *entity.Company) *CompanyActivatedEvent {
	return &CompanyActivatedEvent{
		BaseEvent: events.NewEvent("company.activated", company),
		Company:   company,
	}
}

type CompanyDeactivatedEvent struct {
	*events.BaseEvent
	Company *entity.Company
}

func NewCompanyDeactivatedEvent(company *entity.Company) *CompanyDeactivatedEvent {
	return &CompanyDeactivatedEvent{
		BaseEvent: events.NewEvent("company.deactivated", company),
		Company:   company,
	}
}