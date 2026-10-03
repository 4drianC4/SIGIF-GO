package event

import (
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/events"
)

// CustomerCreatedEvent se publica cuando un nuevo cliente es registrado con éxito.
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

// CustomerUpdatedEvent se publica cuando los datos de un cliente son modificados.
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

// CustomerDeletedEvent se publica cuando un cliente recibe baja lógica.
type CustomerDeletedEvent struct {
	*events.BaseEvent
	CustomerID int64
}

func NewCustomerDeletedEvent(customerID int64) *CustomerDeletedEvent {
	return &CustomerDeletedEvent{
		BaseEvent:  events.NewEvent("customer.deleted", map[string]any{"customer_id": customerID}),
		CustomerID: customerID,
	}
}

// CustomerStatusChangedEvent se publica cuando el estado del cliente cambia.
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
