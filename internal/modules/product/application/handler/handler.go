package handler

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

type CatalogCommandHandler struct {
	service *service.CatalogService
}

func NewCatalogCommandHandler(service *service.CatalogService) *CatalogCommandHandler {
	return &CatalogCommandHandler{service: service}
}

type CatalogQueryHandler struct {
	service *service.CatalogService
}

func NewCatalogQueryHandler(service *service.CatalogService) *CatalogQueryHandler {
	return &CatalogQueryHandler{service: service}
}
