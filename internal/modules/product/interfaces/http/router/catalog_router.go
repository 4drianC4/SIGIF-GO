package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/product/interfaces/http/handler"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/middleware"
)

// CatalogManagerRoles son los roles que pueden registrar productos y categorías.
var CatalogManagerRoles = []string{
	string(userEntity.RoleSuperAdmin),
	string(userEntity.RoleTenantAdmin),
	string(userEntity.RoleCompanyAdmin),
	string(userEntity.RoleManager),
	string(userEntity.RoleInventory),
}

func RegisterCatalogRoutes(router fiber.Router, h *handler.CatalogHTTPHandler) {
	canManageCatalog := middleware.RequireRoles(CatalogManagerRoles...)

	router.Post("/categories", canManageCatalog, h.CreateCategory)
	router.Post("/products", canManageCatalog, h.CreateProduct)
}
