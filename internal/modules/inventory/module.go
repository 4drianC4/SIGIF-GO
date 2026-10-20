package inventory

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/inventory/application/adapter"
	"github.com/sigif/sigif-go/internal/modules/inventory/application/handler"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/service"
	"github.com/sigif/sigif-go/internal/modules/inventory/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/inventory/interfaces/http/handler"
	httpRouter "github.com/sigif/sigif-go/internal/modules/inventory/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewBranchLocationGormRepository, fx.As(new(repository.BranchLocationRepository))),
		fx.Annotate(gorm.NewStockBalanceGormRepository, fx.As(new(repository.StockBalanceRepository))),
		fx.Annotate(gorm.NewMovementGormRepository, fx.As(new(repository.MovementRepository))),
		fx.Annotate(adapter.NewProductReaderAdapter, fx.As(new(repository.ProductReader))),
		service.NewInventoryService,
		handler.NewInventoryCommandHandler,
		handler.NewInventoryQueryHandler,
	),

	fx.Provide(httpHandler.NewInventoryHTTPHandler),

	fx.Invoke(func(app *fiber.App, handler *httpHandler.InventoryHTTPHandler, checker middleware.PermissionChecker) {
		httpRouter.RegisterInventoryRoutes(app.Group("/api/v1"), handler, checker)
	}),
)
