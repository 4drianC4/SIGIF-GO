package handler

import (
	"github.com/sigif/sigif-go/internal/modules/auth/domain/service"
)

type AuthCommandHandler struct {
	service *service.AuthService
}

func NewAuthCommandHandler(service *service.AuthService) *AuthCommandHandler {
	return &AuthCommandHandler{service: service}
}
