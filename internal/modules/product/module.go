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
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(gorm.NewProductGormRepository, fx.As(new(repository.ProductRepository))),
		fx.Annotate(gorm.NewCategoryGormRepository, fx.As(new(repository.CategoryRepository))),
		service.NewCatalogService,
		handler.NewCatalogCommandHandler,
	),

	// Handler HTTP
	fx.Provide(httpHandler.NewCatalogHTTPHandler),

	// Rutas bajo /api/v1 (POST /api/v1/categories, POST /api/v1/products)
	fx.Invoke(func(app *fiber.App, handler *httpHandler.CatalogHTTPHandler) {
		httpRouter.RegisterCatalogRoutes(app.Group("/api/v1"), handler)
	}),
)
