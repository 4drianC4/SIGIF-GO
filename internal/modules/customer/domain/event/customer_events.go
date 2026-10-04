package event

import (
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/events"
)

type CustomerCreatedEvent struct {
	*events.BaseEvent
	Customer *entity.Customer
}

func NewCustomerCreatedEvent(customer *entity.Customer) *CustomerCreatedEvent {
	return &CustomerCreatedEvent{
		BaseEvent: events.NewEvent("customer.created", customer),
		Customer:  customer,
	}
}

type CustomerUpdatedEvent struct {
	*events.BaseEvent
	Customer *entity.Customer
}

func NewCustomerUpdatedEvent(customer *entity.Customer) *CustomerUpdatedEvent {
	return &CustomerUpdatedEvent{
		BaseEvent: events.NewEvent("customer.updated", customer),
		Customer:  customer,
	}
}

type CustomerDeletedEvent struct {
	*events.BaseEvent
	CustomerID uuid.UUID
}

func NewCustomerDeletedEvent(customerID uuid.UUID) *CustomerDeletedEvent {
	return &CustomerDeletedEvent{
		BaseEvent:  events.NewEvent("customer.deleted", map[string]any{"customer_id": customerID}),
		CustomerID: customerID,
	}
}

type CustomerStatusChangedEvent struct {
	*events.BaseEvent
	Customer *entity.Customer
}

func NewCustomerStatusChangedEvent(customer *entity.Customer) *CustomerStatusChangedEvent {
	return &CustomerStatusChangedEvent{
		BaseEvent: events.NewEvent("customer.status_changed", customer),
		Customer:  customer,
	}
}
