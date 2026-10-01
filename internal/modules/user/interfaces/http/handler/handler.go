package handler

import (
	"github.com/sigif/sigif-go/internal/modules/user/application/handler"
)

// UserHTTPHandler expone los endpoints HTTP de usuarios.
type UserHTTPHandler struct {
	cmdHandler   *handler.UserCommandHandler
	queryHandler *handler.UserQueryHandler
}

func NewUserHTTPHandler(cmdHandler *handler.UserCommandHandler, queryHandler *handler.UserQueryHandler) *UserHTTPHandler {
	return &UserHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
	}
}