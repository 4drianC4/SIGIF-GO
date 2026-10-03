package handler

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/shared/events"
)

// CatalogCommandHandler coordina los casos de uso de escritura del catálogo.
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
