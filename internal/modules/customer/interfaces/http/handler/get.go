package handler

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/application/query"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
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
		return response.Error(c, fiber.StatusBadRequest, err)
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

// List searches the company's customers (HU-08-02). Without a status filter
// only active customers are returned; status=all includes every status.
func (h *CustomerHTTPHandler) List(c *fiber.Ctx) error {
	var req dtos.ListCustomersRequest
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

	var statusFilter *entity.CustomerStatus
	switch req.Status {
	case dtos.StatusAll:
	case "":
		st := entity.CustomerStatusActive
		statusFilter = &st
	default:
		st := dtos.StatusFromRequest(req.Status)
		statusFilter = &st
	}

	companyID, err := companyIDFromContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	customers, total, err := h.queryHandler.HandleList(c.UserContext(), query.ListCustomers{
		CompanyID: companyID,
		Q:         req.Q,
		Status:    statusFilter,
		SortBy:    repository.SortField(req.SortBy),
		SortOrder: repository.SortOrder(req.SortOrder),
		Offset:    (page - 1) * limit,
		Limit:     limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	return response.Paginated(c, dto.SummaryFromEntityList(customers), p)
}
