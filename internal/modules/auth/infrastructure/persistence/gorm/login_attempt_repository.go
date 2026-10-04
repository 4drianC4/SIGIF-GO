package gorm

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/mapper"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type LoginAttemptGormRepository struct {
	db *sharedDatabase.Database
}

func NewLoginAttemptGormRepository(db *sharedDatabase.Database) repository.LoginAttemptRepository {
	return &LoginAttemptGormRepository{db: db}
}

func (r *LoginAttemptGormRepository) Create(ctx context.Context, attempt *entity.LoginAttempt) error {
	return r.db.GetDB(ctx).Create(mapper.LoginAttemptToModel(attempt)).Error
}
