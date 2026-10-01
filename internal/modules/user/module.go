package user

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/modules/user/application/port"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/router"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewUserGormRepository, fx.As(new(repository.UserRepository))),
		service.NewUserService,
		handler.NewUserCommandHandler,
		handler.NewUserQueryHandler,
		httpHandler.NewUserHTTPHandler,
	),
	fx.Invoke(registerRoutes),
)

func registerRoutes(app *fiber.App, h *httpHandler.UserHTTPHandler) {
	router.RegisterUserRoutes(app, h)
}

var (
	_ repository.UserRepository = (*gorm.UserGormRepository)(nil)
	_ port.UserCommandPort      = (*handler.UserCommandHandler)(nil)
	_ port.UserQueryPort        = (*handler.UserQueryHandler)(nil)
)