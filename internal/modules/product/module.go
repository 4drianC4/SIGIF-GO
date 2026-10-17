package product

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/product/application/handler"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/product/interfaces/http/handler"
	httpRouter "github.com/sigif/sigif-go/internal/modules/product/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewProductGormRepository, fx.As(new(repository.ProductRepository))),
		fx.Annotate(gorm.NewCategoryGormRepository, fx.As(new(repository.CategoryRepository))),
		fx.Annotate(gorm.NewUnitOfMeasureGormRepository, fx.As(new(repository.UnitOfMeasureRepository))),
		fx.Annotate(gorm.NewTaxGormRepository, fx.As(new(repository.TaxRepository))),
		service.NewCatalogService,
		handler.NewCatalogCommandHandler,
		handler.NewCatalogQueryHandler,
	),

	fx.Provide(httpHandler.NewCatalogHTTPHandler),

	fx.Invoke(func(app *fiber.App, handler *httpHandler.CatalogHTTPHandler, checker middleware.PermissionChecker) {
		httpRouter.RegisterCatalogRoutes(app.Group("/api/v1"), handler, checker)
	}),
)
