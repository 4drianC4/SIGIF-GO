package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/application/query"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *CustomerHTTPHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid customer id", 400))
	}

	companyID, err := companyIDFromContext(c)
	if err != nil {
		return err
	}

	customer, err := h.queryHandler.HandleGet(c.UserContext(), query.GetCustomer{
		ID:        id,
		CompanyID: companyID,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if customer == nil {
		return response.Error(c, fiber.StatusNotFound, sharedErrors.ErrNotFound)
	}

	return response.Success(c, dto.FromEntity(customer))
}

func (h *CustomerHTTPHandler) List(c *fiber.Ctx) error {
	companyID, err := companyIDFromContext(c)
	if err != nil {
		return err
	}

	defaultLimit := h.cfg.Pagination.DefaultLimit
	if defaultLimit < 1 {
		defaultLimit = 20
	}
	maxLimit := h.cfg.Pagination.MaxLimit
	if maxLimit < 1 {
		maxLimit = 100
	}

	page, limit := pagination.Parse(c, defaultLimit, maxLimit)

	qParam := c.Query("q")
	var statusFilter *entity.CustomerStatus
	if s := c.Query("status"); s != "" {
		st := dtos.StatusFromRequest(s)
		if st.IsValid() {
			statusFilter = &st
		}
	}

	customers, total, err := h.queryHandler.HandleList(c.UserContext(), query.ListCustomers{
		CompanyID: companyID,
		Q:         qParam,
		Status:    statusFilter,
		Offset:    (page - 1) * limit,
		Limit:     limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	return response.Paginated(c, dto.FromEntityList(customers), p)
}
