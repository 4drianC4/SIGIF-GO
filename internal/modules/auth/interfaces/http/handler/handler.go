package handler

import (
	"github.com/sigif/sigif-go/internal/modules/auth/application/handler"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type AuthHTTPHandler struct {
	cmdHandler *handler.AuthCommandHandler
	validator  *validator.Validator
}

func NewAuthHTTPHandler(cmdHandler *handler.AuthCommandHandler, validator *validator.Validator) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		cmdHandler: cmdHandler,
		validator:  validator,
	}
}
