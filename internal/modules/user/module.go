package user

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
	httpRouter "github.com/sigif/sigif-go/internal/modules/user/interfaces/http/router"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewUserGormRepository, fx.As(new(repository.UserRepository))),
		service.NewUserService,
		handler.NewUserCommandHandler,
		handler.NewUserQueryHandler,
	),

	// Handler individual
	fx.Provide(httpHandler.NewUserHTTPHandler),

	// routes
	fx.Invoke(func(app *fiber.App, handler *httpHandler.UserHTTPHandler) {
		httpRouter.RegisterUserRoutes(app, handler)
	}),
)