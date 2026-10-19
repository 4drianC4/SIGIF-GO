package supplier

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/handler"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/service"
	"github.com/sigif/sigif-go/internal/modules/supplier/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/handler"
	httpRouter "github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewSupplierGormRepository, fx.As(new(repository.SupplierRepository))),
		fx.Annotate(gorm.NewProductSupplierGormRepository, fx.As(new(repository.ProductSupplierRepository))),
		service.NewSupplierService,
		handler.NewSupplierCommandHandler,
		handler.NewSupplierQueryHandler,
		httpHandler.NewSupplierHTTPHandler,
	),
	fx.Invoke(func(app *fiber.App, h *httpHandler.SupplierHTTPHandler, checker middleware.PermissionChecker) {
		httpRouter.RegisterSupplierRoutes(app.Group("/api/v1"), h, checker)
	}),
)
