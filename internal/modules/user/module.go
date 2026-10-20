package user

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/company"
	"github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/gorm"
	httpHandler "github.com/sigif/sigif-go/internal/modules/user/interfaces/http/handler"
	httpRouter "github.com/sigif/sigif-go/internal/modules/user/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

var Module = fx.Options(
	// User management requires the company catalog and its selector routes.
	company.Module,
	fx.Provide(
		fx.Annotate(gorm.NewUserGormRepository, fx.As(new(repository.UserRepository))),
		fx.Annotate(gorm.NewRoleGormRepository, fx.As(new(repository.RoleRepository))),
		fx.Annotate(gorm.NewPermissionGormRepository, fx.As(new(repository.PermissionRepository))),
		service.NewUserService,
		// Expose the same UserService as the RBAC PermissionChecker used by the
		// shared middleware.
		func(s *service.UserService) middleware.PermissionChecker { return s },
		handler.NewUserCommandHandler,
		handler.NewUserQueryHandler,
	),

	fx.Provide(httpHandler.NewUserHTTPHandler),
	fx.Provide(httpHandler.NewPermissionHTTPHandler),
	fx.Provide(httpHandler.NewRoleHTTPHandler),

	fx.Invoke(func(app *fiber.App, handler *httpHandler.UserHTTPHandler, checker middleware.PermissionChecker) {
		httpRouter.RegisterUserRoutes(app.Group("/api/v1"), handler, checker)
	}),

	fx.Invoke(func(app *fiber.App, permissionHandler *httpHandler.PermissionHTTPHandler, checker middleware.PermissionChecker) {
		httpRouter.RegisterPermissionRoutes(app.Group("/api/v1"), permissionHandler, checker)
	}),

	fx.Invoke(func(app *fiber.App, roleHandler *httpHandler.RoleHTTPHandler, checker middleware.PermissionChecker) {
		httpRouter.RegisterRoleRoutes(app.Group("/api/v1"), roleHandler, checker)
	}),
)
