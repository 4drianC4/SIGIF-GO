package auth

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/auth/application/adapter"
	"github.com/sigif/sigif-go/internal/modules/auth/application/handler"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	authService "github.com/sigif/sigif-go/internal/modules/auth/domain/service"
	"github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/handler"
	httpRouter "github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/router"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewRefreshTokenGormRepository, fx.As(new(repository.RefreshTokenRepository))),
		adapter.NewUserRepoAdapter,
		authService.NewAuthService,
		handler.NewAuthCommandHandler,
	),

	// Handler HTTP individual
	fx.Provide(httpHandler.NewAuthHTTPHandler),

	// Registrar rutas con el router de Fiber
	fx.Invoke(func(app *fiber.App, handler *httpHandler.AuthHTTPHandler) {
		httpRouter.RegisterAuthRoutes(app, handler)
	}),
)