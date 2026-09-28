package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"

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
		fx.Invoke(registerHooks),
		fx.Invoke(startServer),
	)

	app.Run()
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

func startServer(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger, eventBus *events.Bus, jwtManager *jwt.JWTManager) {
	fiberApp := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.App.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.App.IdleTimeout) * time.Second,
		ErrorHandler: middleware.ErrorHandler(),
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})

	fiberApp.Use(recover.New())
	fiberApp.Use(requestid.New())
	fiberApp.Use(fiberLogger.New(fiberLogger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} | ${path} | ${ip} | ${requestid}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))
	fiberApp.Use(helmet.New())
	fiberApp.Use(compress.New())
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.CORS.AllowOrigins, ","),
		AllowMethods:     strings.Join(cfg.CORS.AllowMethods, ","),
		AllowHeaders:     strings.Join(cfg.CORS.AllowHeaders, ","),
		AllowCredentials: cfg.CORS.AllowCredentials,
		ExposeHeaders:    strings.Join(cfg.CORS.ExposeHeaders, ","),
	}))

	if cfg.RateLimit.Enabled {
		fiberApp.Use(limiter.New(limiter.Config{
			Max:        cfg.RateLimit.MaxRequests,
			Expiration: time.Duration(cfg.RateLimit.WindowSeconds) * time.Second,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
		}))
	}

	fiberApp.Use(middleware.RequestID())
	fiberApp.Use(middleware.TenantContext())

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	api := fiberApp.Group("/api/v1")

	api.Use(func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenant_id")
		if tenantID == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "X-Tenant-ID header is required",
			})
		}
		return c.Next()
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := cfg.App.Host + ":" + strconv.Itoa(cfg.App.Port)
			log.Info("Starting HTTP server", zap.String("address", addr))
			go func() {
				if err := fiberApp.Listen(addr); err != nil {
					log.Fatal("HTTP server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping HTTP server...")
			return fiberApp.ShutdownWithContext(ctx)
		},
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	_ = fiberApp.ShutdownWithContext(context.Background())
}