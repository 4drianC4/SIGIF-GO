package handler

import (
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
	"github.com/sigif/sigif-go/internal/shared/events"
)

// CustomerCommandHandler coordina los casos de uso de escritura.
type CustomerCommandHandler struct {
	service  *service.CustomerService
	eventBus *events.Bus
}

func NewCustomerCommandHandler(service *service.CustomerService, eventBus *events.Bus) *CustomerCommandHandler {
	return &CustomerCommandHandler{
		service:  service,
		eventBus: eventBus,
	}
}

// CustomerQueryHandler coordina los casos de uso de lectura.
type CustomerQueryHandler struct {
	service *service.CustomerService
}

func NewCustomerQueryHandler(service *service.CustomerService) *CustomerQueryHandler {
	return &CustomerQueryHandler{
		service: service,
	}
}
