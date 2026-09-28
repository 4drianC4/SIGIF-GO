package company

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"github.com/sigif/sigif-go/internal/modules/company/application/handler"
	"github.com/sigif/sigif-go/internal/modules/company/application/port"
	"github.com/sigif/sigif-go/internal/modules/company/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/company/domain/service"
	"github.com/sigif/sigif-go/internal/modules/company/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/company/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/company/interfaces/http/router"
)

var Module = fx.Options(
	fx.Provide(
		gorm.NewCompanyGormRepository,
		service.NewCompanyService,
		handler.NewCompanyCommandHandler,
		handler.NewCompanyQueryHandler,
		httpHandler.NewCompanyHTTPHandler,
	),
	fx.Invoke(registerRoutes),
)

func registerRoutes(
	app *fiber.App,
	h *httpHandler.CompanyHTTPHandler,
) {
	router.RegisterCompanyRoutes(app, h)
}

var (
	_ repository.CompanyRepository = (*gorm.CompanyGormRepository)(nil)
	_ port.CompanyCommandPort      = (*handler.CompanyCommandHandler)(nil)
	_ port.CompanyQueryPort        = (*handler.CompanyQueryHandler)(nil)
)