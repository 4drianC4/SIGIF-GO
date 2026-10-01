package adapter

import (
	"context"

	"github.com/google/uuid"

	authRepo "github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	userService "github.com/sigif/sigif-go/internal/modules/user/domain/service"
)

// userRepoAdapter adapta el UserService al puerto UserRepo de auth.
type userRepoAdapter struct {
	userService *userService.UserService
}

func NewUserRepoAdapter(userService *userService.UserService) authRepo.UserRepo {
	return &userRepoAdapter{userService: userService}
}

func (a *userRepoAdapter) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*userEntity.User, error) {
	return a.userService.GetByEmail(ctx, tenantID, email)
}

func (a *userRepoAdapter) GetByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error) {
	return a.userService.GetByID(ctx, id)
}

func (a *userRepoAdapter) RecordLogin(ctx context.Context, userID uuid.UUID) error {
	return a.userService.RecordLogin(ctx, userID)
}

var _ authRepo.UserRepo = (*userRepoAdapter)(nil)
