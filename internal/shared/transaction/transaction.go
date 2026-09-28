package transaction

import (
	"context"
	"gorm.io/gorm"
)

type TransactionManager interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

func (m *GormTransactionManager) Begin(ctx context.Context) (context.Context, error) {
	tx := m.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return context.WithValue(ctx, txKey{}, tx), nil
}

func (m *GormTransactionManager) Commit(ctx context.Context) error {
	tx := GetTx(ctx)
	if tx == nil {
		return nil
	}
	return tx.Commit().Error
}

func (m *GormTransactionManager) Rollback(ctx context.Context) error {
	tx := GetTx(ctx)
	if tx == nil {
		return nil
	}
	return tx.Rollback().Error
}

func (m *GormTransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		newCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(newCtx)
	})
}

type txKey struct{}

func GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}