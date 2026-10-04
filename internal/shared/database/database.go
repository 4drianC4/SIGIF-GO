package database

import (
	"context"
	"fmt"
	"github.com/sigif/sigif-go/internal/shared/config"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(lc fx.Lifecycle, cfg *config.Config) (*Database, error) {
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
		Logger: logger.Default.LogMode(getLogLevel(cfg.Database.LogLevel)),
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return sqlDB.Close()
		},
	})

	log.Println("Database connected successfully")
	return &Database{DB: db}, nil
}

func getLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Silent
	}
}

var Module = fx.Options(
	fx.Provide(NewDatabase),
)

type TxKey struct{}

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, TxKey{}, tx)
}

func GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(TxKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}

func (d *Database) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return d.DB.WithContext(ctx).Transaction(fn)
}

func (d *Database) GetDB(ctx context.Context) *gorm.DB {
	if tx := GetTx(ctx); tx != nil {
		return tx
	}
	return d.DB.WithContext(ctx)
}
