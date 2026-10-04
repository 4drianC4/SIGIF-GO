package repository

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
)

// LoginAttemptRepository is the persistence port for login attempts.
type LoginAttemptRepository interface {
	Create(ctx context.Context, attempt *entity.LoginAttempt) error
}
