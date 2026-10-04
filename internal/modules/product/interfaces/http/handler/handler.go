package handler

import (
	"github.com/sigif/sigif-go/internal/modules/product/application/handler"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type CatalogHTTPHandler struct {
	cmdHandler *handler.CatalogCommandHandler
	validator  *validator.Validator
}

func NewCatalogHTTPHandler(cmdHandler *handler.CatalogCommandHandler, val *validator.Validator) *CatalogHTTPHandler {
	return &CatalogHTTPHandler{
		cmdHandler: cmdHandler,
		validator:  val,
	}
}
