package gorm

import (
	"context"

	gormlib "gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type Transactor struct {
	db *sharedDatabase.Database
}

func NewTransactor(db *sharedDatabase.Database) repository.Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if sharedDatabase.GetTx(ctx) != nil {
		return fn(ctx)
	}
	return t.db.Transaction(ctx, func(tx *gormlib.DB) error {
		return fn(sharedDatabase.WithTx(ctx, tx))
	})
}
