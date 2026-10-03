package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/application/query"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// GetByID maneja GET /customers/:id.
func (h *CustomerHTTPHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid customer id", 400))
	}

	tenantID := dtos.TenantIDFromContext(c)

	customer, err := h.queryHandler.HandleGet(c.UserContext(), query.GetCustomer{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if customer == nil {
		return response.Error(c, fiber.StatusNotFound, sharedErrors.ErrNotFound)
	}

	customerDTO := dto.FromEntity(customer)
	return response.Success(c, dtos.ToResponse(customerDTO))
}

// List maneja GET /customers.
func (h *CustomerHTTPHandler) List(c *fiber.Ctx) error {
	tenantID := dtos.TenantIDFromContext(c)

	page, limit := pagination.Parse(c, 20, 100)

	qParam := c.Query("q")
	var statusFilter *entity.CustomerStatus
	if s := c.Query("status"); s != "" {
		st := dtos.StatusFromRequest(s)
		if st.IsValid() {
			statusFilter = &st
		}
	}

	customers, total, err := h.queryHandler.HandleList(c.UserContext(), query.ListCustomers{
		TenantID: tenantID,
		Q:        qParam,
		Status:   statusFilter,
		Offset:   (page - 1) * limit,
		Limit:    limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	dtoList := dto.FromEntityList(customers)
	return response.Paginated(c, dtos.ToResponseList(dtoList), p)
}
