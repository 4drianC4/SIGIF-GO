package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/cashregister/application/dto"
	"github.com/sigif/sigif-go/internal/modules/cashregister/application/query"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// ListCashRegisters lists the active cash registers of the company with their
// availability (HU-10-01, HU-10-02). Without a filter every cash register is
// returned.
func (h *CashRegisterHTTPHandler) ListCashRegisters(c *fiber.Ctx) error {
	var req dtos.ListCashRegistersRequest
	if err := c.QueryParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid query parameters", 400))
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

	var availability *entity.Availability
	if req.Availability != "" && req.Availability != dtos.AvailabilityAll {
		a := entity.Availability(req.Availability)
		availability = &a
	}

	_, companyID, err := actorFromContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	registers, total, err := h.queryHandler.HandleList(c.UserContext(), query.ListCashRegisters{
		CompanyID:    companyID,
		Availability: availability,
		Offset:       (page - 1) * limit,
		Limit:        limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Paginated(c, dto.FromRegisterViews(registers), pagination.New(page, limit, total))
}
