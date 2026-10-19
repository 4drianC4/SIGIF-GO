package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/supplier/application/command"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/dto"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/query"
	"github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *SupplierHTTPHandler) AssociateProduct(c *fiber.Ctx) error {
	companyID, supplierID, err := identity(c)
	if err != nil {
		return fail(c, err)
	}
	var req dtos.AssociateProductRequest
	if err := h.body(c, &req); err != nil {
		return fail(c, err)
	}
	a, err := h.commands.HandleAssociate(c.UserContext(), command.AssociateProduct{
		CompanyID: companyID, SupplierID: supplierID, ProductID: uuid.MustParse(req.ProductID),
		SupplierCode: req.SupplierCode, LeadTimeDays: req.LeadTimeDays, CurrentCost: req.CurrentCost, Preferred: req.Preferred})
	if err != nil {
		return fail(c, err)
	}
	return response.Created(c, dto.FromAssociation(a))
}

func (h *SupplierHTTPHandler) AssociateProducts(c *fiber.Ctx) error {
	companyID, supplierID, err := identity(c)
	if err != nil {
		return fail(c, err)
	}
	var req dtos.AssociateProductsRequest
	if err := h.body(c, &req); err != nil {
		return fail(c, err)
	}
	ids := make([]uuid.UUID, len(req.ProductIDs))
	for i, id := range req.ProductIDs {
		ids[i] = uuid.MustParse(id)
	}
	items, err := h.commands.HandleAssociateProducts(c.UserContext(), command.AssociateProducts{CompanyID: companyID, SupplierID: supplierID, ProductIDs: ids})
	if err != nil {
		return fail(c, err)
	}
	return response.Created(c, dto.FromAssociations(items))
}

func (h *SupplierHTTPHandler) UpdateAssociation(c *fiber.Ctx) error {
	companyID, supplierID, err := identity(c)
	if err != nil {
		return fail(c, err)
	}
	productID, err := pathUUID(c, "product_id")
	if err != nil {
		return fail(c, err)
	}
	var req dtos.UpdateAssociationRequest
	if err := h.body(c, &req); err != nil {
		return fail(c, err)
	}
	a, err := h.commands.HandleUpdateAssociation(c.UserContext(), command.UpdateProductAssociation{
		CompanyID: companyID, SupplierID: supplierID, ProductID: productID, SupplierCode: req.SupplierCode,
		LeadTimeDays: req.LeadTimeDays, CurrentCost: req.CurrentCost, Preferred: req.Preferred})
	if err != nil {
		return fail(c, err)
	}
	return response.Success(c, dto.FromAssociation(a))
}

func (h *SupplierHTTPHandler) UnlinkProduct(c *fiber.Ctx) error {
	companyID, supplierID, err := identity(c)
	if err != nil {
		return fail(c, err)
	}
	productID, err := pathUUID(c, "product_id")
	if err != nil {
		return fail(c, err)
	}
	if err := h.commands.HandleUnlink(c.UserContext(), command.UnlinkProduct{CompanyID: companyID, SupplierID: supplierID, ProductID: productID}); err != nil {
		return fail(c, err)
	}
	return response.NoContent(c)
}

func (h *SupplierHTTPHandler) ProductOptions(c *fiber.Ctx) error {
	companyID, supplierID, err := identity(c)
	if err != nil {
		return fail(c, err)
	}
	var q dtos.ProductOptionsQuery
	if err := h.query(c, &q); err != nil {
		return fail(c, err)
	}
	page, limit := pagination.Parse(c, 20, 100)
	items, total, err := h.queries.HandleProductOptions(c.UserContext(), query.ProductOptions{CompanyID: companyID, SupplierID: supplierID, Filter: q.Filter(), Offset: (page - 1) * limit, Limit: limit})
	if err != nil {
		return fail(c, err)
	}
	return response.Paginated(c, dto.FromProductOptions(items), pagination.New(page, limit, total))
}

func (h *SupplierHTTPHandler) ListProductSuppliers(c *fiber.Ctx) error {
	return h.productSuppliers(c, false)
}
func (h *SupplierHTTPHandler) SupplierOptions(c *fiber.Ctx) error {
	return h.productSuppliers(c, true)
}
func (h *SupplierHTTPHandler) productSuppliers(c *fiber.Ctx, options bool) error {
	companyID, err := company(c)
	if err != nil {
		return fail(c, err)
	}
	productID, err := pathUUID(c, "product_id")
	if err != nil {
		return fail(c, err)
	}
	var q dtos.ProductsQuery
	if err := h.query(c, &q); err != nil {
		return fail(c, err)
	}
	page, limit := pagination.Parse(c, 20, 100)
	if options {
		items, total, err := h.queries.HandleSupplierOptions(c.UserContext(), query.SupplierOptions{CompanyID: companyID, ProductID: productID, Search: q.Search, Offset: (page - 1) * limit, Limit: limit})
		if err != nil {
			return fail(c, err)
		}
		return response.Paginated(c, dto.FromSupplierOptions(items), pagination.New(page, limit, total))
	}
	items, total, err := h.queries.HandleListProductSuppliers(c.UserContext(), query.ListProductSuppliers{CompanyID: companyID, ProductID: productID, Search: q.Search, Offset: (page - 1) * limit, Limit: limit})
	if err != nil {
		return fail(c, err)
	}
	return response.Paginated(c, dto.FromProductSuppliers(items), pagination.New(page, limit, total))
}
