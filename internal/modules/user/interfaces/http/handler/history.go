package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/dto"
	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *UserHTTPHandler) History(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
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
	page, limit := pagination.Parse(c, defaultLimit, maxLimit)

	history, total, err := h.queryHandler.HandleListHistory(c.UserContext(), query.ListUserHistory{
		UserID: id,
		Offset: (page - 1) * limit,
		Limit:  limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Paginated(c, dto.FromHistoryList(history), pagination.New(page, limit, total))
}
