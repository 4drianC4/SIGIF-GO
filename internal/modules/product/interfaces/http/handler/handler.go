package handler

import (
	"github.com/sigif/sigif-go/internal/modules/product/application/handler"
	sharedValidator "github.com/sigif/sigif-go/internal/shared/validator"
)

// CatalogHTTPHandler expone los endpoints HTTP de productos y categorías.
type CatalogHTTPHandler struct {
	cmdHandler *handler.CatalogCommandHandler
	validator  *sharedValidator.Validator
}

func NewCatalogHTTPHandler(cmdHandler *handler.CatalogCommandHandler) *CatalogHTTPHandler {
	return &CatalogHTTPHandler{
		cmdHandler: cmdHandler,
		validator:  sharedValidator.New(),
	}
}
