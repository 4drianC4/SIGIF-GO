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
	var req dtos.RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	user, err := h.cmdHandler.HandleRegister(c.UserContext(), command.RegisterUser{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		CompanyID: &req.CompanyID,
		RoleName:  req.Role,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	c.Set("Cache-Control", "no-store")
	return response.Created(c, dtos.RegisterUserResponse{User: dtos.ToResponse(user.User), GeneratedPassword: user.GeneratedPassword})
}

func (h *UserHTTPHandler) List(c *fiber.Ctx) error {
	defaultLimit := h.cfg.Pagination.DefaultLimit
	if defaultLimit < 1 {
		defaultLimit = 20
	}
	maxLimit := h.cfg.Pagination.MaxLimit
	if maxLimit < 1 {
		maxLimit = 100
	}

	page, limit := pagination.Parse(c, defaultLimit, maxLimit)

	users, total, err := h.queryHandler.HandleList(c.UserContext(), querypkg.ListUsers{
		Offset: (page - 1) * limit,
		Limit:  limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	return response.Paginated(c, dtos.ToResponseList(users), p)
}
