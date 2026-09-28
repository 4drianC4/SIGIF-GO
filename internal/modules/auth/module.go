package auth

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"github.com/sigif/sigif-go/internal/modules/auth/application/handler"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/service"
	"github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/router"
)

var Module = fx.Options(
	fx.Provide(
		gorm.NewRefreshTokenGormRepository,
		service.NewAuthService,
		handler.NewAuthCommandHandler,
		httpHandler.NewAuthHTTPHandler,
	),
	fx.Invoke(registerRoutes),
)

func registerRoutes(
	app *fiber.App,
	h *httpHandler.AuthHTTPHandler,
) {
	router.RegisterAuthRoutes(app, h)
}

var (
	_ repository.RefreshTokenRepository = (*gorm.RefreshTokenGormRepository)(nil)
)