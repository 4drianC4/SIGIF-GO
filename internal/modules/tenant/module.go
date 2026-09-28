package tenant

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"github.com/sigif/sigif-go/internal/modules/tenant/application/handler"
	"github.com/sigif/sigif-go/internal/modules/tenant/application/port"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/service"
	"github.com/sigif/sigif-go/internal/modules/tenant/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/tenant/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/tenant/interfaces/http/router"
)

var Module = fx.Options(
	fx.Provide(
		gorm.NewTenantGormRepository,
		service.NewTenantService,
		handler.NewTenantCommandHandler,
		handler.NewTenantQueryHandler,
		httpHandler.NewTenantHTTPHandler,
	),
	fx.Invoke(registerRoutes),
)

func registerRoutes(
	app *fiber.App,
	h *httpHandler.TenantHTTPHandler,
) {
	router.RegisterTenantRoutes(app, h)
}

var (
	_ repository.TenantRepository = (*gorm.TenantGormRepository)(nil)
	_ port.TenantCommandPort      = (*handler.TenantCommandHandler)(nil)
	_ port.TenantQueryPort        = (*handler.TenantQueryHandler)(nil)
)