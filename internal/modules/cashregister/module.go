package cashregister

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/adapter"
	"github.com/sigif/sigif-go/internal/modules/cashregister/application/handler"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
	"github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/cashregister/interfaces/http/handler"
	httpRouter "github.com/sigif/sigif-go/internal/modules/cashregister/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

// Module wires the cash register assignment slice (RF-10). The cashier names
// come from the user module through an adapter.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewCashRegisterGormRepository, fx.As(new(repository.CashRegisterRepository))),
		adapter.NewCashierDirectoryAdapter,
		service.NewCashRegisterService,
		handler.NewCashRegisterCommandHandler,
		handler.NewCashRegisterQueryHandler,
	),

	fx.Provide(httpHandler.NewCashRegisterHTTPHandler),

	fx.Invoke(func(app *fiber.App, h *httpHandler.CashRegisterHTTPHandler, checker middleware.PermissionChecker) {
		httpRouter.RegisterCashRegisterRoutes(app.Group("/api/v1"), h, checker)
	}),
)
