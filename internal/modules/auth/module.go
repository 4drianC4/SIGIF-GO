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
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewSessionGormRepository, fx.As(new(repository.SessionRepository))),
		fx.Annotate(gorm.NewLoginAttemptGormRepository, fx.As(new(repository.LoginAttemptRepository))),
		adapter.NewUserRepoAdapter,
		authService.NewAuthService,
		handler.NewAuthCommandHandler,
		// Expose the session repository as the middleware SessionReader.
		func(r repository.SessionRepository) middleware.SessionReader { return r },
	),

	fx.Provide(httpHandler.NewAuthHTTPHandler),

	fx.Invoke(func(app *fiber.App, handler *httpHandler.AuthHTTPHandler) {
		httpRouter.RegisterAuthRoutes(app.Group("/api/v1"), handler)
	}),
)
