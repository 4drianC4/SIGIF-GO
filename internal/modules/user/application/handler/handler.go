package handler

import (
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	"github.com/sigif/sigif-go/internal/shared/events"
)

// UserCommandHandler coordina los casos de uso de escritura.
type UserCommandHandler struct {
	service  *service.UserService
	eventBus *events.Bus
}

func NewUserCommandHandler(service *service.UserService, eventBus *events.Bus) *UserCommandHandler {
	return &UserCommandHandler{
		service:  service,
		eventBus: eventBus,
	}
}

// UserQueryHandler coordina los casos de uso de lectura.
type UserQueryHandler struct {
	service *service.UserService
}

func NewUserQueryHandler(service *service.UserService) *UserQueryHandler {
	return &UserQueryHandler{service: service}
}
