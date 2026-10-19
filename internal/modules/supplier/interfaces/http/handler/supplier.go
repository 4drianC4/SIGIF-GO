package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/command"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/dto"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/query"
	"github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
	"strings"
)

func (h *SupplierHTTPHandler) Create(c *fiber.Ctx) error {
	companyID, err := company(c)
	if err != nil {
		return fail(c, err)
	}
	var req dtos.SupplierRequest
	if err := h.body(c, &req); err != nil {
		return fail(c, err)
	}
	created, err := h.commands.HandleCreate(c.UserContext(), command.CreateSupplier{CompanyID: companyID, Fields: req.Fields()})
	if err != nil {
		return fail(c, err)
	}
	return response.Created(c, dto.FromSupplier(created))
}

func (h *SupplierHTTPHandler) Update(c *fiber.Ctx) error {
	companyID, id, err := identity(c)
	if err != nil {
		return fail(c, err)
	}
	var req dtos.SupplierRequest
	if err := h.body(c, &req); err != nil {
		return fail(c, err)
	}
	updated, err := h.commands.HandleUpdate(c.UserContext(), command.UpdateSupplier{CompanyID: companyID, ID: id, Fields: req.Fields()})
	if err != nil {
		return fail(c, err)
	}
	return response.Success(c, dto.FromSupplier(updated))
}

func (h *SupplierHTTPHandler) Deactivate(c *fiber.Ctx) error {
	companyID, id, err := identity(c)
	if err != nil {
		return fail(c, err)
	}
	updated, err := h.commands.HandleDeactivate(c.UserContext(), command.DeactivateSupplier{CompanyID: companyID, ID: id})
	if err != nil {
		return fail(c, err)
	}
	return response.Success(c, dto.FromSupplier(updated))
}

func (h *SupplierHTTPHandler) List(c *fiber.Ctx) error {
	companyID, err := company(c)
	if err != nil {
		return fail(c, err)
	}
	var q dtos.ListQuery
	if err := h.query(c, &q); err != nil {
		return fail(c, err)
	}
	q.Search, q.MainIndustry = strings.TrimSpace(q.Search), strings.TrimSpace(q.MainIndustry)
	page, limit := pagination.Parse(c, 20, 100)
	items, total, err := h.queries.HandleList(c.UserContext(), query.ListSuppliers{CompanyID: companyID, Filter: q.Filter(), Offset: (page - 1) * limit, Limit: limit})
	if err != nil {
		return fail(c, err)
	}
	return response.Paginated(c, dto.FromList(items), pagination.New(page, limit, total))
}

func (h *SupplierHTTPHandler) Get(c *fiber.Ctx) error {
	companyID, id, err := identity(c)
	if err != nil {
		return fail(c, err)
	}
	result, err := h.queries.HandleGet(c.UserContext(), query.GetSupplier{CompanyID: companyID, ID: id})
	if err != nil {
		return fail(c, err)
	}
	return response.Success(c, dto.FromSupplier(result))
}

func (h *SupplierHTTPHandler) Summary(c *fiber.Ctx) error {
	companyID, err := company(c)
	if err != nil {
		return fail(c, err)
	}
	result, err := h.queries.HandleSummary(c.UserContext(), query.SupplierSummary{CompanyID: companyID})
	if err != nil {
		return fail(c, err)
	}
	return response.Success(c, dto.FromSummary(result))
}

func (h *SupplierHTTPHandler) FormOptions(c *fiber.Ctx) error {
	companyID, err := company(c)
	if err != nil {
		return fail(c, err)
	}
	options, err := h.queries.HandleFormOptions(c.UserContext(), companyID)
	if err != nil {
		return fail(c, err)
	}
	return response.Success(c, dto.FormOptions{Industries: options.Industries, PaymentTerms: options.PaymentTerms})
}
