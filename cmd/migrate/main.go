package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	authModel "github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/model"
	companyModel "github.com/sigif/sigif-go/internal/modules/company/infrastructure/persistence/model"
	customerModel "github.com/sigif/sigif-go/internal/modules/customer/infrastructure/persistence/model"
	userModel "github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/seed"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/database/migrations"
	sharedLogger "github.com/sigif/sigif-go/internal/shared/logger"
)

var rootCmd = &cobra.Command{
	Use:   "sigif-migrate",
	Short: "SIGIF Database Migration Tool",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		run(runMigrations)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show applied and pending versioned migrations",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		run(showStatus)
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the last versioned migration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		run(rollbackLast)
	},
}

var forceCmd = &cobra.Command{
	Use:   "force VERSION",
	Short: "Set the migration version without running SQL (clears a dirty state)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("VERSION must be a number: %w", err)
		}
		run(func(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := migrations.Force(cfg.Database.DSN(), version); err != nil {
						log.Error("Force failed", zap.Error(err))
						return err
					}
					if err := printStatus(cfg); err != nil {
						return err
					}
					os.Exit(0)
					return nil
				},
			})
		})
		return nil
	},
}

func run(invoke any) {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			sharedLogger.NewLogger,
			database.NewDatabase,
		),
		fx.Invoke(invoke),
	)
	app.Run()
}

func runMigrations(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger, db *database.Database) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Running database migrations...")

			if err := autoMigrate(db.DB); err != nil {
				log.Error("Migration failed", zap.Error(err))
				return err
			}

			log.Info("Applying versioned migrations...")
			if err := migrations.Up(cfg.Database.DSN()); err != nil {
				log.Error("Versioned migration failed", zap.Error(err))
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

func showStatus(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := printStatus(cfg); err != nil {
				log.Error("Status failed", zap.Error(err))
				return err
			}
			os.Exit(0)
			return nil
		},
	})
}

func rollbackLast(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := migrations.Down(cfg.Database.DSN()); err != nil {
				log.Error("Rollback failed", zap.Error(err))
				return err
			}
			if err := printStatus(cfg); err != nil {
				return err
			}
			os.Exit(0)
			return nil
		},
	})
}

func printStatus(cfg *config.Config) error {
	list, err := migrations.Status(cfg.Database.DSN())
	if err != nil {
		return err
	}
	for _, m := range list {
		state := "pending"
		if m.Applied {
			state = "applied"
		}
		fmt.Printf("%06d  %-8s %s\n", m.Version, state, m.Name)
	}
	return nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&companyModel.CompanyModel{},
		&userModel.UserModel{},
		&userModel.RoleModel{},
		&userModel.PermissionModel{},
		&userModel.RolePermissionModel{},
		&authModel.SessionModel{},
		&authModel.LoginAttemptModel{},
		&customerModel.CustomerModel{},
	)
}

func main() {
	rootCmd.AddCommand(statusCmd, downCmd, forceCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
