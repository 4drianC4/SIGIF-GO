package container

import (
	"go.uber.org/fx"

	"github.com/sigif/sigif-go/internal/modules/auth"
	"github.com/sigif/sigif-go/internal/modules/product"
	"github.com/sigif/sigif-go/internal/modules/user"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/events"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	sharedLogger "github.com/sigif/sigif-go/internal/shared/logger"
)

var Module = fx.Options(
	fx.Provide(
		config.NewConfig,
		sharedLogger.NewLogger,
		database.NewDatabase,
		events.NewBus,
		jwt.NewManager,
		fx.Annotate(clock.NewRealClock, fx.As(new(clock.Clock))),
	),
	user.Module,
	auth.Module,
	product.Module,
)