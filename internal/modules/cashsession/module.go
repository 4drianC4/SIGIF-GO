package cashsession

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/cashsession/application/adapter"
	"github.com/sigif/sigif-go/internal/modules/cashsession/application/handler"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
	"github.com/sigif/sigif-go/internal/modules/cashsession/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/cashsession/interfaces/http/handler"
	httpRouter "github.com/sigif/sigif-go/internal/modules/cashsession/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

// Module wires the cash session slice (RF-09). The cash register assignments
// come from the cash register module (RF-10) through an adapter, and it
// provides the middleware.CashSessionChecker that guards cash transactions.
var Module = fx.Options(
	fx.Provide(
		gorm.NewCashSessionGormRepository,
		adapter.NewRegisterAccessAdapter,
		adapter.NewSessionCheckerAdapter,
		service.NewCashSessionService,
		handler.NewCashSessionCommandHandler,
		handler.NewCashSessionQueryHandler,
	),

	fx.Provide(httpHandler.NewCashSessionHTTPHandler),

	fx.Invoke(func(app *fiber.App, h *httpHandler.CashSessionHTTPHandler, checker middleware.PermissionChecker) {
		httpRouter.RegisterCashSessionRoutes(app.Group("/api/v1"), h, checker)
	}),
)
