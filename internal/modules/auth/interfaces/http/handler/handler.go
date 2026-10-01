package handler

import (
	"github.com/sigif/sigif-go/internal/modules/auth/application/handler"
)

type AuthHTTPHandler struct {
	cmdHandler *handler.AuthCommandHandler
}

func NewAuthHTTPHandler(cmdHandler *handler.AuthCommandHandler) *AuthHTTPHandler {
	return &AuthHTTPHandler{cmdHandler: cmdHandler}
}
