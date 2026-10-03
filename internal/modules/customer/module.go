package customer

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/customer/application/handler"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
	"github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/handler"
	httpRouter "github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/router"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewCustomerGormRepository, fx.As(new(repository.CustomerRepository))),
		service.NewCustomerService,
		handler.NewCustomerCommandHandler,
		handler.NewCustomerQueryHandler,
	),

	fx.Provide(httpHandler.NewCustomerHTTPHandler),

	fx.Invoke(func(app *fiber.App, h *httpHandler.CustomerHTTPHandler) {
		httpRouter.RegisterCustomerRoutes(app.Group("/api/v1"), h)
	}),
)
