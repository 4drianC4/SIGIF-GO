package handler

import (
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
)

type CustomerCommandHandler struct {
	service *service.CustomerService
}

func NewCustomerCommandHandler(service *service.CustomerService) *CustomerCommandHandler {
	return &CustomerCommandHandler{service: service}
}

type CustomerQueryHandler struct {
	service *service.CustomerService
}

func NewCustomerQueryHandler(service *service.CustomerService) *CustomerQueryHandler {
	return &CustomerQueryHandler{service: service}
}
