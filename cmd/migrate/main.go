package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	userModel "github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	authModel "github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/model"
	productModel "github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	sharedLogger "github.com/sigif/sigif-go/internal/shared/logger"
)

var rootCmd = &cobra.Command{
	Use:   "sigif-migrate",
	Short: "SIGIF Database Migration Tool",
	Run: func(cmd *cobra.Command, args []string) {
		app := fx.New(
			fx.Provide(
				config.NewConfig,
				sharedLogger.NewLogger,
				database.NewDatabase,
			),
			fx.Invoke(runMigrations),
		)
		app.Run()
	},
}

func runMigrations(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger, db *database.Database) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Running database migrations...")

			if err := autoMigrate(db.DB); err != nil {
				log.Fatal("Migration failed", zap.Error(err))
			}

			log.Info("Migrations completed successfully")
			os.Exit(0)
			return nil
		},
	})
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&userModel.UserModel{},
		&authModel.RefreshTokenModel{},
		&productModel.CategoryModel{},
		&productModel.ProductModel{},
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}