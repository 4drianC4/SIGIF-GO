package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/application/dto"
	"github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/config"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
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

// Update edits the description of a permission (PATCH /permissions/:id).
func (h *PermissionHTTPHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}

	var req dtos.UpdatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	permission, err := h.cmdHandler.HandleUpdatePermission(c.UserContext(), command.UpdatePermission{
		ID:          id,
		Description: req.Description,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.PermissionFromEntity(permission))
}

// Delete removes a permission (DELETE /permissions/:id). System permissions and
// permissions still assigned to a role are rejected with 409.
func (h *PermissionHTTPHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}

	if err := h.cmdHandler.HandleDeletePermission(c.UserContext(), command.DeletePermission{ID: id}); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

// Export streams the permission catalog as CSV (GET /permissions/export),
// honoring the optional module filter.
func (h *PermissionHTTPHandler) Export(c *fiber.Ctx) error {
	var req dtos.ExportPermissionsRequest
	if err := c.QueryParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	permissions, err := h.queryHandler.HandleExportPermissions(c.UserContext(), query.ExportPermissions{
		Module: req.Module,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF") // UTF-8 BOM so spreadsheet tools read accents.
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "module", "operation", "code", "description", "is_system"})
	for _, permission := range permissions {
		_ = writer.Write([]string{
			permission.ID.String(),
			permission.Module,
			permission.Operation,
			permission.Key(),
			permission.Description,
			strconv.FormatBool(entity.IsSystemPermission(permission.Module, permission.Operation)),
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	filename := "permissions.csv"
	if module := strings.ToLower(strings.TrimSpace(req.Module)); module != "" {
		filename = fmt.Sprintf("permissions_%s.csv", module)
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	return c.Status(fiber.StatusOK).Send(buf.Bytes())
}
