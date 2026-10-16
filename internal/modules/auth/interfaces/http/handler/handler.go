package handler

import (
	"github.com/sigif/sigif-go/internal/modules/auth/application/handler"
	"github.com/sigif/sigif-go/internal/shared/authcookie"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type AuthHTTPHandler struct {
	cmdHandler *handler.AuthCommandHandler
	validator  *validator.Validator
	cookie     authcookie.Settings
}

func NewAuthHTTPHandler(
	cmdHandler *handler.AuthCommandHandler,
	validator *validator.Validator,
	cookie authcookie.Settings,
) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		cmdHandler: cmdHandler,
		validator:  validator,
		cookie:     cookie,
	}
}
