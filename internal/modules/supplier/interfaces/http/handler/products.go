package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/dto"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/query"
	"github.com/sigif/sigif-go/internal/modules/supplier/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *SupplierHTTPHandler) ListProducts(c *fiber.Ctx) error {
	companyID, id, err := identity(c)
	if err != nil {
		return fail(c, err)
	}
	var q dtos.ProductsQuery
	if err := h.query(c, &q); err != nil {
		return fail(c, err)
	}
	page, limit := pagination.Parse(c, 20, 100)
	items, total, err := h.queries.HandleListProducts(c.UserContext(), query.ListSupplierProducts{
		CompanyID: companyID, SupplierID: id, Search: q.Search, Offset: (page - 1) * limit, Limit: limit})
	if err != nil {
		return fail(c, err)
	}
	return response.Paginated(c, dto.FromProducts(items), pagination.New(page, limit, total))
}
