package handler

import (
	"github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

// UserHTTPHandler exposes the user HTTP endpoints.
type UserHTTPHandler struct {
	cmdHandler   *handler.UserCommandHandler
	queryHandler *handler.UserQueryHandler
	validator    *validator.Validator
	cfg          *config.Config
}

func NewUserHTTPHandler(
	cmdHandler *handler.UserCommandHandler,
	queryHandler *handler.UserQueryHandler,
	validator *validator.Validator,
	cfg *config.Config,
) *UserHTTPHandler {
	return &UserHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
		validator:    validator,
		cfg:          cfg,
	}
}
