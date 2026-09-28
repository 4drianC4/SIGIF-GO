package container

import (
	"go.uber.org/fx"
	"github.com/sigif/sigif-go/internal/modules/auth"
	"github.com/sigif/sigif-go/internal/modules/company"
	"github.com/sigif/sigif-go/internal/modules/dashboard"
	"github.com/sigif/sigif-go/internal/modules/hardware"
	"github.com/sigif/sigif-go/internal/modules/inventory"
	"github.com/sigif/sigif-go/internal/modules/minimarket"
	"github.com/sigif/sigif-go/internal/modules/pharmacy"
	"github.com/sigif/sigif-go/internal/modules/product"
	"github.com/sigif/sigif-go/internal/modules/sales"
	"github.com/sigif/sigif-go/internal/modules/tenant"
	"github.com/sigif/sigif-go/internal/modules/user"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/events"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/logger"
)

var Module = fx.Options(
	fx.Provide(
		config.NewConfig,
		logger.NewLogger,
		database.NewDatabase,
		events.NewBus,
		jwt.NewManager,
	),
	tenant.Module,
	company.Module,
	user.Module,
	auth.Module,
	product.Module,
	inventory.Module,
	sales.Module,
	pharmacy.Module,
	minimarket.Module,
	hardware.Module,
	dashboard.Module,
)