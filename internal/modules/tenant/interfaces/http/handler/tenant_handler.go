package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/tenant/application/command"
	"github.com/sigif/sigif-go/internal/modules/tenant/application/handler"
	"github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/tenant/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

type TenantHTTPHandler struct {
	cmdHandler   *handler.TenantCommandHandler
	queryHandler *handler.TenantQueryHandler
}

func NewTenantHTTPHandler(cmdHandler *handler.TenantCommandHandler, queryHandler *handler.TenantQueryHandler) *TenantHTTPHandler {
	return &TenantHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
	}
}

func (h *TenantHTTPHandler) Create(c *fiber.Ctx) error {
	var req dtos.CreateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	businessType := entity.BusinessType(req.BusinessType)
	cmd := command.CreateTenantCommand{
		Name:         req.Name,
		Slug:         req.Slug,
		BusinessType: businessType,
	}

	tenant, err := h.cmdHandler.HandleCreate(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Created(c, dtos.ToTenantResponse(tenant))
}

func (h *TenantHTTPHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	tenant, err := h.queryHandler.GetByID(c.UserContext(), id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if tenant == nil {
		return response.Error(c, fiber.StatusNotFound, err)
	}

	return response.Success(c, dtos.ToTenantResponse(tenant))
}

func (h *TenantHTTPHandler) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")

	tenant, err := h.queryHandler.GetBySlug(c.UserContext(), slug)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if tenant == nil {
		return response.Error(c, fiber.StatusNotFound, err)
	}

	return response.Success(c, dtos.ToTenantResponse(tenant))
}

func (h *TenantHTTPHandler) List(c *fiber.Ctx) error {
	page, limit := pagination.Parse(c, 20, 100)

	tenants, total, err := h.queryHandler.List(c.UserContext(), (page-1)*limit, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	return response.Paginated(c, dtos.ToTenantResponseList(tenants), p)
}

func (h *TenantHTTPHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	var req dtos.UpdateTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.UpdateTenantCommand{
		ID:       id,
		Name:     req.Name,
		Settings: req.Settings,
	}

	tenant, err := h.cmdHandler.HandleUpdate(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dtos.ToTenantResponse(tenant))
}

func (h *TenantHTTPHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.DeleteTenantCommand{ID: id}
	if err := h.cmdHandler.HandleDelete(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *TenantHTTPHandler) Activate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.ActivateTenantCommand{ID: id}
	if err := h.cmdHandler.HandleActivate(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *TenantHTTPHandler) Deactivate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.DeactivateTenantCommand{ID: id}
	if err := h.cmdHandler.HandleDeactivate(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}