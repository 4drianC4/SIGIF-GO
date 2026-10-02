package main

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/sigif/sigif-go/internal/modules/auth"
	"github.com/sigif/sigif-go/internal/modules/user"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/events"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	sharedLogger "github.com/sigif/sigif-go/internal/shared/logger"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			sharedLogger.NewLogger,
			database.NewDatabase,
			events.NewBus,
			jwt.NewManager,
			newFiberApp,
			fx.Annotate(clock.NewRealClock, fx.As(new(clock.Clock))),
		),
		user.Module,
		auth.Module,
		fx.Invoke(registerHooks),
		fx.Invoke(startServer),
	)

	app.Run()
}

func newFiberApp(cfg *config.Config, jwtManager *jwt.JWTManager) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.App.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.App.IdleTimeout) * time.Second,
		ErrorHandler: middleware.ErrorHandler(),
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} | ${path} | ${ip} | ${requestid}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))
	app.Use(helmet.New())
	app.Use(compress.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.CORS.AllowOrigins, ","),
		AllowMethods:     strings.Join(cfg.CORS.AllowMethods, ","),
		AllowHeaders:     strings.Join(cfg.CORS.AllowHeaders, ","),
		AllowCredentials: cfg.CORS.AllowCredentials,
		ExposeHeaders:    strings.Join(cfg.CORS.ExposeHeaders, ","),
	}))

	if cfg.RateLimit.Enabled {
		app.Use(limiter.New(limiter.Config{
			Max:        cfg.RateLimit.MaxRequests,
			Expiration: time.Duration(cfg.RateLimit.WindowSeconds) * time.Second,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
		}))
	}

	app.Use(middleware.RequestID())
	app.Use(middleware.TenantContext())
	app.Use(middleware.AuthRequired(jwtManager))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	return app
}

func registerHooks(lc fx.Lifecycle, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Application starting...")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Application stopping...")
			return nil
		},
	})
}

func startServer(lc fx.Lifecycle, cfg *config.Config, app *fiber.App, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := cfg.App.Host + ":" + strconv.Itoa(cfg.App.Port)
			log.Info("Starting HTTP server", zap.String("address", addr))

			go func() {
				if err := app.Listen(addr); err != nil {
					log.Error("HTTP server error", zap.Error(err))
					_ = app.ShutdownWithContext(ctx)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping HTTP server...")
			return app.ShutdownWithContext(ctx)
		},
	})
}
