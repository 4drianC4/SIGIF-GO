package handler

import (
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
)

type CashSessionCommandHandler struct {
	service *service.CashSessionService
}

func NewCashSessionCommandHandler(service *service.CashSessionService) *CashSessionCommandHandler {
	return &CashSessionCommandHandler{service: service}
}

type CashSessionQueryHandler struct {
	service *service.CashSessionService
}

func NewCashSessionQueryHandler(service *service.CashSessionService) *CashSessionQueryHandler {
	return &CashSessionQueryHandler{service: service}
}
