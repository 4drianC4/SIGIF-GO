package handler

import (
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
)

type CashRegisterCommandHandler struct {
	service *service.CashRegisterService
}

func NewCashRegisterCommandHandler(service *service.CashRegisterService) *CashRegisterCommandHandler {
	return &CashRegisterCommandHandler{service: service}
}

type CashRegisterQueryHandler struct {
	service *service.CashRegisterService
}

func NewCashRegisterQueryHandler(service *service.CashRegisterService) *CashRegisterQueryHandler {
	return &CashRegisterQueryHandler{service: service}
}
