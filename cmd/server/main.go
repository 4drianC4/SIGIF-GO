package main

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/sigif/sigif-go/internal/modules/customer"
	"github.com/sigif/sigif-go/internal/modules/inventory"
	"github.com/sigif/sigif-go/internal/modules/product"
	"github.com/sigif/sigif-go/internal/modules/supplier"
	"github.com/sigif/sigif-go/internal/modules/user"
	"github.com/sigif/sigif-go/internal/shared/authcookie"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	sharedLogger "github.com/sigif/sigif-go/internal/shared/logger"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			sharedLogger.NewLogger,
			database.NewDatabase,
			jwt.NewManager,
			authcookie.New,
			validator.New,
			newFiberApp,
			fx.Annotate(clock.NewRealClock, fx.As(new(clock.Clock))),
		),
		user.Module,
		auth.Module,
		customer.Module,
		product.Module,
		inventory.Module,
		supplier.Module,
		fx.Invoke(registerHooks),
		fx.Invoke(startServer),
	)

	app.Run()
}

func newFiberApp(
	cfg *config.Config,
	jwtManager *jwt.JWTManager,
	sessions middleware.SessionReader,
	authCookie authcookie.Settings,
) (*fiber.App, error) {
	if err := validateCORS(cfg.CORS); err != nil {
		return nil, err
	}

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
	// CSRF: the session cookie is sent automatically by the browser, so
	// state-changing requests must come from an allowed origin.
	app.Use(middleware.OriginGuard(cfg.CORS.AllowOrigins))
	app.Use(middleware.AuthRequired(jwtManager, sessions, authCookie))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	return app, nil
}

// validateCORS rejects a wildcard origin when credentials are allowed: the
// auth cookie must only be sent to explicitly trusted origins.
func validateCORS(cfg config.CORSConfig) error {
	if !cfg.AllowCredentials {
		return nil
	}
	for _, origin := range cfg.AllowOrigins {
		if strings.TrimSpace(origin) == "*" {
			return errors.New("cors.allow_origins cannot contain \"*\" when cors.allow_credentials is true")
		}
	}
	return nil
}

func registerHooks(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-super-secret-jwt-key-change-in-production" {
				log.Warn("JWT secret is using the default placeholder value; set SIGIF_JWT_SECRET in production")
			}
			if !cfg.AuthCookie.Secure && cfg.App.Env == "production" {
				log.Warn("auth cookie is not Secure in production; set SIGIF_AUTH_COOKIE_SECURE=true behind HTTPS")
			}
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
