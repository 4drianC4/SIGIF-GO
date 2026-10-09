package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
)

// UserCommandHandler coordinates write use cases.
type UserCommandHandler struct {
	service *service.UserService
	history *service.HistoryService
	tx      repository.Transactor
}

func NewUserCommandHandler(service *service.UserService, history *service.HistoryService, tx repository.Transactor) *UserCommandHandler {
	return &UserCommandHandler{service: service, history: history, tx: tx}
}

// UserQueryHandler coordinates read use cases.
type UserQueryHandler struct {
	service *service.UserService
	history *service.HistoryService
}

func NewUserQueryHandler(service *service.UserService, history *service.HistoryService) *UserQueryHandler {
	return &UserQueryHandler{service: service, history: history}
}

func (h *UserCommandHandler) HandleRegister(ctx context.Context, cmd command.RegisterUser) (*service.RegisteredUser, error) {
	var registered *service.RegisteredUser
	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		result, err := h.service.Register(ctx, service.RegisterUserInput{
			CompanyID: cmd.CompanyID,
			RoleName:  cmd.RoleName,
			FirstName: cmd.FirstName,
			LastName:  cmd.LastName,
			Email:     cmd.Email,
		})
		if err != nil {
			return err
		}
		registered = result
		return h.record(ctx, result.User.ID, entity.HistoryActionCreated, nil)
	})
	if err != nil {
		return nil, err
	}
	return registered, nil
}

func (h *UserCommandHandler) HandleEdit(ctx context.Context, cmd command.EditUser) (*entity.AppUser, error) {
	var edited *entity.AppUser
	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		current, err := h.service.GetByID(ctx, cmd.ID)
		if err != nil {
			return err
		}
		before := *current

		user, err := h.service.Edit(ctx, cmd.ID, service.EditUserInput{
			CompanyID: cmd.CompanyID,
			FirstName: cmd.FirstName,
			LastName:  cmd.LastName,
			Email:     cmd.Email,
			Password:  cmd.Password,
			RoleName:  cmd.RoleName,
		})
		if err != nil {
			return err
		}
		edited = user

		if changes := userChanges(&before, user); len(changes) > 0 {
			if err := h.record(ctx, user.ID, entity.HistoryActionUpdated, changes); err != nil {
				return err
			}
		}
		if cmd.Password != nil {
			return h.record(ctx, user.ID, entity.HistoryActionPasswordChanged, nil)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return edited, nil
}

func (h *UserCommandHandler) HandleChangePassword(ctx context.Context, cmd command.ChangePassword) error {
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := h.service.ChangePassword(ctx, cmd.ID, cmd.CurrentPassword, cmd.NewPassword); err != nil {
			return err
		}
		return h.record(ctx, cmd.ID, entity.HistoryActionPasswordChanged, nil)
	})
}

func (h *UserCommandHandler) HandleDelete(ctx context.Context, cmd command.DeleteUser) error {
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := h.service.Delete(ctx, cmd.ID); err != nil {
			return err
		}
		return h.record(ctx, cmd.ID, entity.HistoryActionDeleted, nil)
	})
}

func (h *UserCommandHandler) HandleChangeStatus(ctx context.Context, cmd command.ChangeStatus) error {
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if cmd.Active {
			if err := h.service.Activate(ctx, cmd.ID); err != nil {
				return err
			}
			return h.record(ctx, cmd.ID, entity.HistoryActionActivated, nil)
		}
		if err := h.service.Deactivate(ctx, cmd.ID); err != nil {
			return err
		}
		return h.record(ctx, cmd.ID, entity.HistoryActionDeactivated, nil)
	})
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

func (h *UserQueryHandler) HandleListHistory(ctx context.Context, q query.ListUserHistory) ([]*entity.UserHistory, int64, error) {
	return h.history.List(ctx, q.UserID, q.Offset, q.Limit)
}
