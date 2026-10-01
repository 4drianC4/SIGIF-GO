package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	querypkg "github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *UserHTTPHandler) Create(c *fiber.Ctx) error {
	var req dtos.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	tenantID := dtos.TenantIDFromContext(c)
	cmd := command.CreateUser{
		TenantID:  tenantID,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Roles:     req.Roles,
	}

	user, err := h.cmdHandler.HandleCreate(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Created(c, dtos.ToResponse(user))
}

func (h *UserHTTPHandler) List(c *fiber.Ctx) error {
	tenantID := dtos.TenantIDFromContext(c)

	page, limit := pagination.Parse(c, 20, 100)

	users, total, err := h.queryHandler.HandleList(c.UserContext(), querypkg.ListUsers{
		TenantID: tenantID,
		Offset:   (page - 1) * limit,
		Limit:    limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	return response.Paginated(c, dtos.ToResponseList(users), p)
}