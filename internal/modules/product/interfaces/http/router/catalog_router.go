package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/shared/config"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
)

const (
	defaultIdentifierMaxRequests   = 30
	defaultIdentifierWindowSeconds = 60
)

func RegisterCatalogRoutes(router fiber.Router, h *handler.CatalogHTTPHandler, checker middleware.PermissionChecker, rateLimit config.IdentifierRateLimits) {
	products := router.Group("/products")
	identifierLimiter := newIdentifierLimiter(rateLimit)
	products.Get("/", middleware.RequirePermission(checker, "products", "list"), h.ListProducts)
	products.Get("/summary", middleware.RequirePermission(checker, "products", "list"), h.ProductSummary)
	products.Get("/validate-duplicate", middleware.RequirePermission(checker, "products", "create"), identifierLimiter, h.ValidateDuplicate)
	products.Get("/verificar-identificador", middleware.RequirePermission(checker, "products", "create"), identifierLimiter, h.VerifyIdentifier)
	products.Post("/", middleware.RequirePermission(checker, "products", "create"), h.CreateProduct)
	products.Get("/:id", middleware.RequirePermission(checker, "products", "read"), h.GetProduct)
	products.Put("/:id", middleware.RequirePermission(checker, "products", "update"), h.UpdateProduct)
	products.Patch("/:id/price", middleware.RequirePermission(checker, "products", "update"), h.UpdateProductPrice)
	products.Patch("/:id/status", middleware.RequirePermission(checker, "products", "update"), h.SetProductStatus)
	products.Get("/:id/price-history", middleware.RequirePermission(checker, "products", "read"), h.GetPriceHistory)
	products.Delete("/:id", middleware.RequirePermission(checker, "products", "delete"), h.DeleteProduct)

	categories := router.Group("/categories")
	categories.Get("/", middleware.RequirePermission(checker, "categories", "list"), h.ListCategories)
	categories.Post("/", middleware.RequirePermission(checker, "categories", "create"), h.CreateCategory)

	router.Get("/units-of-measure", h.ListUnitsOfMeasure)
	router.Get("/taxes", h.ListTaxes)
}

// newIdentifierLimiter builds the per-user rate limit shared by the identifier
// lookups. Missing or non-positive settings fall back to the HU-06-02 default.
func newIdentifierLimiter(rateLimit config.IdentifierRateLimits) fiber.Handler {
	maxRequests := rateLimit.MaxRequests
	if maxRequests <= 0 {
		maxRequests = defaultIdentifierMaxRequests
	}
	windowSeconds := rateLimit.WindowSeconds
	if windowSeconds <= 0 {
		windowSeconds = defaultIdentifierWindowSeconds
	}
	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: time.Duration(windowSeconds) * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			if userID, ok := middleware.UserIDFromContext(c.UserContext()); ok {
				return userID.String()
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusTooManyRequests, sharedErrors.ErrTooManyRequests)
		},
	})
}
