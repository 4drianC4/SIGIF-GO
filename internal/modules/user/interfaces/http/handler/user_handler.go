package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/application/handler"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
)

type UserHTTPHandler struct {
	cmdHandler   *handler.UserCommandHandler
	queryHandler *handler.UserQueryHandler
}

func NewUserHTTPHandler(cmdHandler *handler.UserCommandHandler, queryHandler *handler.UserQueryHandler) *UserHTTPHandler {
	return &UserHTTPHandler{
		cmdHandler:   cmdHandler,
		queryHandler: queryHandler,
	}
}

func (h *UserHTTPHandler) Create(c *fiber.Ctx) error {
	var req dtos.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	tenantID := c.Locals("tenant_id").(uuid.UUID)
	cmd := command.CreateUserCommand{
		TenantID:  tenantID,
		CompanyID: req.CompanyID,
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

	return response.Created(c, dtos.ToUserResponse(user))
}

func (h *UserHTTPHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	user, err := h.queryHandler.GetByID(c.UserContext(), id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if user == nil {
		return response.Error(c, fiber.StatusNotFound, err)
	}

	return response.Success(c, dtos.ToUserResponse(user))
}

func (h *UserHTTPHandler) GetByEmail(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return response.Error(c, fiber.StatusBadRequest, nil)
	}

	user, err := h.queryHandler.GetByEmail(c.UserContext(), email)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	if user == nil {
		return response.Error(c, fiber.StatusNotFound, err)
	}

	return response.Success(c, dtos.ToUserResponse(user))
}

func (h *UserHTTPHandler) List(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uuid.UUID)
	companyID := c.Query("company_id")
	var companyIDPtr *uuid.UUID
	if companyID != "" {
		id, err := uuid.Parse(companyID)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, err)
		}
		companyIDPtr = &id
	}

	page, limit := pagination.Parse(c, 20, 100)

	var users []*entity.User
	var total int64
	var err error

	if companyIDPtr != nil {
		users, err = h.queryHandler.GetByCompanyID(c.UserContext(), *companyIDPtr)
		total = int64(len(users))
	} else {
		users, total, err = h.queryHandler.List(c.UserContext(), tenantID, (page-1)*limit, limit)
	}

	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	p := pagination.New(page, limit, total)
	return response.Paginated(c, dtos.ToUserResponseList(users), p)
}

func (h *UserHTTPHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	var req dtos.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.UpdateUserCommand{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		AvatarURL: req.AvatarURL,
		Roles:     req.Roles,
		Settings:  req.Settings,
	}

	user, err := h.cmdHandler.HandleUpdate(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dtos.ToUserResponse(user))
}

func (h *UserHTTPHandler) ChangePassword(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	var req dtos.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.ChangePasswordCommand{
		ID:              id,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := h.cmdHandler.HandleChangePassword(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *UserHTTPHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.DeleteUserCommand{ID: id}
	if err := h.cmdHandler.HandleDelete(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *UserHTTPHandler) Activate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.ActivateUserCommand{ID: id}
	if err := h.cmdHandler.HandleActivate(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *UserHTTPHandler) Deactivate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.DeactivateUserCommand{ID: id}
	if err := h.cmdHandler.HandleDeactivate(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}

func (h *UserHTTPHandler) Suspend(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.SuspendUserCommand{ID: id}
	if err := h.cmdHandler.HandleSuspend(c.UserContext(), cmd); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.NoContent(c)
}