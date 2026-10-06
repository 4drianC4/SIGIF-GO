package handler

import (
	"github.com/sigif/sigif-go/internal/modules/product/application/handler"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type CatalogHTTPHandler struct {
	cmdHandler   *handler.CatalogCommandHandler
	queryHandler *handler.CatalogQueryHandler
	validator    *validator.Validator
}

func NewCatalogHTTPHandler(
	cmdHandler *handler.CatalogCommandHandler,
	queryHandler *handler.CatalogQueryHandler,
	val *validator.Validator,
) *CatalogHTTPHandler {
	return &CatalogHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
		validator:    val,
	}
}
