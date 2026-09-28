package pagination

import (
	"math"
	"strconv"
	"github.com/gofiber/fiber/v2"
)

type Page struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int `json:"total_pages"`
}

func New(page, limit int, total int64) *Page {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return &Page{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

func (p *Page) Offset() int {
	return (p.Page - 1) * p.Limit
}

func Parse(c *fiber.Ctx, defaultLimit, maxLimit int) (int, int) {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", strconv.Itoa(defaultLimit)))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > maxLimit {
		limit = defaultLimit
	}

	return page, limit
}

type Cursor struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit"`
	HasMore bool   `json:"has_more"`
}

func NewCursor(cursor string, limit int, hasMore bool) *Cursor {
	return &Cursor{
		Cursor:  cursor,
		Limit:   limit,
		HasMore: hasMore,
	}
}