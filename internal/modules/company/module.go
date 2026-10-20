package company

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/company/application/handler"
	"github.com/sigif/sigif-go/internal/modules/company/domain/service"
	"github.com/sigif/sigif-go/internal/modules/company/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/company/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/company/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(gorm.NewCompanyGormRepository, service.NewCompanyService, handler.NewCompanyQueryHandler, httpHandler.NewCompanyHTTPHandler),
	fx.Invoke(func(app *fiber.App, h *httpHandler.CompanyHTTPHandler, checker middleware.PermissionChecker) {
		router.RegisterCompanyRoutes(app.Group("/api/v1"), h, checker)
	}),
)
