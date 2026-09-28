package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/application/command"
	"github.com/sigif/sigif-go/internal/modules/company/application/handler"
	"github.com/sigif/sigif-go/internal/modules/company/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

type CompanyHTTPHandler struct {
	cmdHandler   *handler.CompanyCommandHandler
	queryHandler *handler.CompanyQueryHandler
}

func NewCompanyHTTPHandler(cmdHandler *handler.CompanyCommandHandler, queryHandler *handler.CompanyQueryHandler) *CompanyHTTPHandler {
	return &CompanyHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
	}
}

func (h *CompanyHTTPHandler) Create(c *fiber.Ctx) error {
	var req dtos.CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	cmd := command.CreateCompanyCommand{
		TenantID:  tenantID,
		Name:      req.Name,
		LegalName: req.LegalName,
		TaxID:     req.TaxID,
	}

	company, err := h.cmdHandler.HandleCreate(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Created(c, dtos.ToCompanyResponse(company))
}

func (h *CompanyHTTPHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	company, err := h.queryHandler.GetByID(c.UserContext(), id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if company == nil {
		return response.Error(c, fiber.StatusNotFound, err)
	}

	return response.Success(c, dtos.ToCompanyResponse(company))
}

func (h *CompanyHTTPHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	page, limit := pagination.Parse(c, 20, 100)

	companies, total, err := h.queryHandler.List(c.UserContext(), tenantID, (page-1)*limit, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	return response.Paginated(c, dtos.ToCompanyResponseList(companies), p)
}

func (h *CompanyHTTPHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	var req dtos.UpdateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.UpdateCompanyCommand{
		ID:         id,
		Name:       req.Name,
		LegalName:  req.LegalName,
		TaxID:      req.TaxID,
		Email:      req.Email,
		Phone:      req.Phone,
		Address:    req.Address,
		City:       req.City,
		State:      req.State,
		Country:    req.Country,
		PostalCode: req.PostalCode,
		Settings:   req.Settings,
	}

	company, err := h.cmdHandler.HandleUpdate(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dtos.ToCompanyResponse(company))
}

func (h *CompanyHTTPHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.DeleteCompanyCommand{ID: id}
	if err := h.cmdHandler.HandleDelete(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *CompanyHTTPHandler) Activate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.ActivateCompanyCommand{ID: id}
	if err := h.cmdHandler.HandleActivate(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *CompanyHTTPHandler) Deactivate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.DeactivateCompanyCommand{ID: id}
	if err := h.cmdHandler.HandleDeactivate(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}