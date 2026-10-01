package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/application/query"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// UserCommandPort define las operaciones de escritura sobre usuarios.
type UserCommandPort interface {
	HandleCreate(ctx context.Context, cmd command.CreateUser) (*entity.User, error)
	HandleUpdate(ctx context.Context, cmd command.UpdateUser) (*entity.User, error)
	HandleChangePassword(ctx context.Context, cmd command.ChangePassword) error
	HandleDelete(ctx context.Context, cmd command.DeleteUser) error
	HandleChangeStatus(ctx context.Context, cmd command.ChangeStatus) error
	HandleRecordLogin(ctx context.Context, id uuid.UUID) error
}

// UserQueryPort define las operaciones de lectura sobre usuarios.
type UserQueryPort interface {
	HandleGet(ctx context.Context, q query.GetUser) (*entity.User, error)
	HandleGetByEmail(ctx context.Context, q query.GetUserByEmail) (*entity.User, error)
	HandleList(ctx context.Context, q query.ListUsers) ([]*entity.User, int64, error)
}
