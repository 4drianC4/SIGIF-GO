package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
)

// UserCommandHandler coordinates write use cases.
type UserCommandHandler struct {
	service *service.UserService
}

func NewUserCommandHandler(service *service.UserService) *UserCommandHandler {
	return &UserCommandHandler{service: service}
}

// UserQueryHandler coordinates read use cases.
type UserQueryHandler struct {
	service *service.UserService
}

func NewUserQueryHandler(service *service.UserService) *UserQueryHandler {
	return &UserQueryHandler{service: service}
}

func (h *UserCommandHandler) HandleRegister(ctx context.Context, cmd command.RegisterUser) (*entity.AppUser, error) {
	return h.service.Register(ctx, service.RegisterUserInput{
		CompanyID: cmd.CompanyID,
		RoleName:  cmd.RoleName,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
		Email:     cmd.Email,
		Password:  cmd.Password,
		Area:      cmd.Area,
	})
}

func (h *UserCommandHandler) HandleEdit(ctx context.Context, cmd command.EditUser) (*entity.AppUser, error) {
	return h.service.Edit(ctx, cmd.ID, service.EditUserInput{
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
		Email:     cmd.Email,
		Password:  cmd.Password,
		RoleName:  cmd.RoleName,
		Area:      cmd.Area,
	})
}

func (h *UserCommandHandler) HandleChangePassword(ctx context.Context, cmd command.ChangePassword) error {
	return h.service.ChangePassword(ctx, cmd.ID, cmd.CurrentPassword, cmd.NewPassword)
}

func (h *UserCommandHandler) HandleDelete(ctx context.Context, cmd command.DeleteUser) error {
	return h.service.Delete(ctx, cmd.ID)
}

func (h *UserCommandHandler) HandleChangeStatus(ctx context.Context, cmd command.ChangeStatus) error {
	if cmd.Active {
		return h.service.Activate(ctx, cmd.ID)
	}
	return h.service.Deactivate(ctx, cmd.ID)
}

func (h *UserQueryHandler) HandleGet(ctx context.Context, q query.GetUser) (*entity.AppUser, error) {
	return h.service.GetByID(ctx, q.ID)
}

func (h *UserQueryHandler) HandleGetByEmail(ctx context.Context, q query.GetUserByEmail) (*entity.AppUser, error) {
	return h.service.GetByEmail(ctx, q.Email)
}

func (h *UserQueryHandler) HandleList(ctx context.Context, q query.ListUsers) ([]*entity.AppUser, int64, error) {
	return h.service.List(ctx, q.Offset, q.Limit)
}

func (h *UserCommandHandler) HandleCreatePermission(ctx context.Context, cmd command.CreatePermission) (*entity.Permission, error) {
	return h.service.CreatePermission(ctx, service.CreatePermissionParams{
		Module:      cmd.Module,
		Operation:   cmd.Operation,
		Code:        cmd.Code,
		Description: cmd.Description,
	})
}

func (h *UserQueryHandler) HandleListPermissions(ctx context.Context, q query.ListPermissions) ([]*entity.Permission, int64, error) {
	return h.service.ListPermissions(ctx, q.Module, q.Offset, q.Limit)
}

func (h *UserQueryHandler) HandleListPermissionModules(_ context.Context, _ query.ListPermissionModules) []entity.ModuleWithOperations {
	return h.service.PermissionModules()
}
