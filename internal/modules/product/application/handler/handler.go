package handler

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/shared/events"
)

type CatalogCommandHandler struct {
	service  *service.CatalogService
	eventBus *events.Bus
}

func NewCatalogCommandHandler(service *service.CatalogService, eventBus *events.Bus) *CatalogCommandHandler {
	return &CatalogCommandHandler{
		service:  service,
		eventBus: eventBus,
	}
}

type CatalogQueryHandler struct {
	service *service.CatalogService
}

func NewCatalogQueryHandler(service *service.CatalogService) *CatalogQueryHandler {
	return &CatalogQueryHandler{service: service}
}
