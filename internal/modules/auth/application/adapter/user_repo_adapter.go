package adapter

import (
	"context"

	"github.com/google/uuid"

	authRepo "github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	userService "github.com/sigif/sigif-go/internal/modules/user/domain/service"
)

// userRepoAdapter adapts the user module's UserService to the auth UserRepo port.
type userRepoAdapter struct {
	userService *userService.UserService
}

func NewUserRepoAdapter(userService *userService.UserService) authRepo.UserRepo {
	return &userRepoAdapter{userService: userService}
}

func (a *userRepoAdapter) GetByEmail(ctx context.Context, email string) (*userEntity.AppUser, error) {
	return a.userService.GetByEmail(ctx, email)
}

func (a *userRepoAdapter) GetByID(ctx context.Context, id uuid.UUID) (*userEntity.AppUser, error) {
	return a.userService.GetByID(ctx, id)
}

func (a *userRepoAdapter) RecordAccess(ctx context.Context, userID uuid.UUID) error {
	return a.userService.RecordAccess(ctx, userID)
}

func (a *userRepoAdapter) PermissionsByRole(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	return a.userService.PermissionCodesByRole(ctx, roleID)
}

var _ authRepo.UserRepo = (*userRepoAdapter)(nil)
