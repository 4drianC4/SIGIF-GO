package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/application/dto"
	"github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

// PermissionHTTPHandler exposes the permission catalog endpoints (HU-082-01):
// the list of modules, the registered permissions and the registration of a
// new permission for later assignment to roles.
type PermissionHTTPHandler struct {
	cmdHandler   *handler.UserCommandHandler
	queryHandler *handler.UserQueryHandler
	validator    *validator.Validator
	cfg          *config.Config
}

func NewPermissionHTTPHandler(
	cmdHandler *handler.UserCommandHandler,
	queryHandler *handler.UserQueryHandler,
	val *validator.Validator,
	cfg *config.Config,
) *PermissionHTTPHandler {
	return &PermissionHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
		validator:    val,
		cfg:          cfg,
	}
}

func (h *PermissionHTTPHandler) ListModules(c *fiber.Ctx) error {
	modules := h.queryHandler.HandleListPermissionModules(c.UserContext(), query.ListPermissionModules{})
	return response.Success(c, dto.PermissionModuleListFromEntity(modules))
}

// List returns the registered permissions, optionally filtered by module.
func (h *PermissionHTTPHandler) List(c *fiber.Ctx) error {
	var req dtos.ListPermissionsRequest
	if err := c.QueryParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	defaultLimit := h.cfg.Pagination.DefaultLimit
	if defaultLimit < 1 {
		defaultLimit = 20
	}
	maxLimit := h.cfg.Pagination.MaxLimit
	if maxLimit < 1 {
		maxLimit = 100
	}

	page, limit := 1, defaultLimit
	if req.Page != nil {
		page = *req.Page
	}
	if req.Limit != nil {
		if *req.Limit > maxLimit {
			return response.ValidationError(c, map[string]string{"limit": fmt.Sprintf("limit must be %d or less", maxLimit)})
		}
		limit = *req.Limit
	}

	permissions, total, err := h.queryHandler.HandleListPermissions(c.UserContext(), query.ListPermissions{
		Module: req.Module,
		Offset: (page - 1) * limit,
		Limit:  limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	return response.Paginated(c, dto.PermissionListFromEntity(permissions), p)
}

func (h *PermissionHTTPHandler) Create(c *fiber.Ctx) error {
	var req dtos.CreatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	permission, err := h.cmdHandler.HandleCreatePermission(c.UserContext(), command.CreatePermission{
		Module:      req.Module,
		Operation:   req.Operation,
		Code:        req.Code,
		Description: req.Description,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Created(c, dto.PermissionFromEntity(permission))
}
