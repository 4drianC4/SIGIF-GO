package adapter

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/repository"
	userService "github.com/sigif/sigif-go/internal/modules/user/domain/service"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// cashierDirectoryAdapter adapts the user module's UserService to the
// cashregister CashierDirectory port.
type cashierDirectoryAdapter struct {
	userService *userService.UserService
}

func NewCashierDirectoryAdapter(userService *userService.UserService) repository.CashierDirectory {
	return &cashierDirectoryAdapter{userService: userService}
}

// Names returns the full name of each cashier; a user that no longer exists is
// left out so the caller shows it without a name.
func (a *cashierDirectoryAdapter) Names(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	names := make(map[uuid.UUID]string, len(userIDs))
	for _, id := range userIDs {
		if _, ok := names[id]; ok {
			continue
		}
		user, err := a.userService.GetByID(ctx, id)
		var appErr *sharedErrors.AppError
		if errors.As(err, &appErr) && appErr.Code == sharedErrors.CodeNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		names[id] = strings.TrimSpace(user.FirstName + " " + user.LastName)
	}
	return names, nil
}

var _ repository.CashierDirectory = (*cashierDirectoryAdapter)(nil)
