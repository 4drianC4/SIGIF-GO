package pharmacy

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var Module = fx.Options()

func registerRoutes(app *fiber.App) {
}

var _ = registerRoutes