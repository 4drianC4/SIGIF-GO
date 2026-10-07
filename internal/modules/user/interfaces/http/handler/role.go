package handler

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/dto"
	"github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/config"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

// RoleHTTPHandler exposes the role listing (HU-082-02): read-only, no role is
// created or modified here.
type RoleHTTPHandler struct {
	queryHandler *handler.UserQueryHandler
	validator    *validator.Validator
	cfg          *config.Config
}

func NewRoleHTTPHandler(
	queryHandler *handler.UserQueryHandler,
	val *validator.Validator,
	cfg *config.Config,
) *RoleHTTPHandler {
	return &RoleHTTPHandler{
		queryHandler: queryHandler,
		validator:    val,
		cfg:          cfg,
	}
}

// List returns the roles the caller may see: the ones of their company plus
// the system roles. The company always comes from the token, never from the
// query string, so roles of another company are never listed.
func (h *RoleHTTPHandler) List(c *fiber.Ctx) error {
	var req dtos.ListRolesRequest
	if err := c.QueryParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid query parameters", 400))
	}
	req.Q = strings.TrimSpace(req.Q)
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

	// A caller without a company (e.g. the global superadmin) only sees the
	// system roles.
	var companyID *uuid.UUID
	if id, ok := middleware.CompanyIDFromContext(c.UserContext()); ok {
		companyID = &id
	}

	var typeFilter *entity.RoleType
	if req.Type != "" {
		t := entity.RoleType(req.Type)
		typeFilter = &t
	}

	var statusFilter *entity.RoleStatus
	switch req.Status {
	case "", dtos.StatusAll:
	default:
		st := entity.RoleStatus(req.Status)
		statusFilter = &st
	}

	roles, total, err := h.queryHandler.HandleListRoles(c.UserContext(), query.ListRoles{
		CompanyID: companyID,
		Q:         req.Q,
		Type:      typeFilter,
		Status:    statusFilter,
		Offset:    (page - 1) * limit,
		Limit:     limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	return response.Paginated(c, dto.RoleSummaryListFromEntity(roles), p)
}
