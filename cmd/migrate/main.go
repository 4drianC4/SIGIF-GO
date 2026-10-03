package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	authModel "github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/model"
	userModel "github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/seed"
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
				log.Error("Migration failed", zap.Error(err))
				return err
			}

			log.Info("Seeding roles, permissions and default admin...")
			if err := seed.Seed(ctx, db.DB, cfg); err != nil {
				log.Error("Seeding failed", zap.Error(err))
				return err
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
		&userModel.RoleModel{},
		&userModel.PermissionModel{},
		&userModel.RolePermissionModel{},
		&authModel.SessionModel{},
		&authModel.LoginAttemptModel{},
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
