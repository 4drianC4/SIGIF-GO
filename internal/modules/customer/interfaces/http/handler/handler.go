package handler

import (
	"github.com/sigif/sigif-go/internal/modules/customer/application/handler"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type CustomerHTTPHandler struct {
	cmdHandler   *handler.CustomerCommandHandler
	queryHandler *handler.CustomerQueryHandler
	validator    *validator.Validator
}

func NewCustomerHTTPHandler(
	cmdHandler *handler.CustomerCommandHandler,
	queryHandler *handler.CustomerQueryHandler,
	val *validator.Validator,
) *CustomerHTTPHandler {
	return &CustomerHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
		validator:    val,
	}
}
