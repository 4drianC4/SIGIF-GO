package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	tenantEntity "github.com/sigif/sigif-go/internal/modules/tenant/domain/entity"
	companyEntity "github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	authEntity "github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	productEntity "github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	inventoryEntity "github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	salesEntity "github.com/sigif/sigif-go/internal/modules/sales/domain/entity"
	pharmacyEntity "github.com/sigif/sigif-go/internal/modules/pharmacy/domain/entity"
	minimarketEntity "github.com/sigif/sigif-go/internal/modules/minimarket/domain/entity"
	hardwareEntity "github.com/sigif/sigif-go/internal/modules/hardware/domain/entity"
	dashboardEntity "github.com/sigif/sigif-go/internal/modules/dashboard/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/logger"
)

var rootCmd = &cobra.Command{
	Use:   "sigif-migrate",
	Short: "SIGIF Database Migration Tool",
	Run: func(cmd *cobra.Command, args []string) {
		app := fx.New(
			fx.Provide(
				config.NewConfig,
				logger.NewLogger,
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
		&tenantEntity.Tenant{},
		&companyEntity.Company{},
		&userEntity.User{},
		&authEntity.RefreshToken{},
		&productEntity.Product{},
		&productEntity.Category{},
		&productEntity.Brand{},
		&productEntity.Unit{},
		&productEntity.Tax{},
		&inventoryEntity.StockMovement{},
		&inventoryEntity.Warehouse{},
		&inventoryEntity.Batch{},
		&inventoryEntity.Supplier{},
		&inventoryEntity.PurchaseOrder{},
		&inventoryEntity.PurchaseOrderItem{},
		&salesEntity.Sale{},
		&salesEntity.SaleItem{},
		&salesEntity.Customer{},
		&salesEntity.Payment{},
		&salesEntity.Return{},
		&salesEntity.ReturnItem{},
		&pharmacyEntity.Prescription{},
		&pharmacyEntity.PrescriptionItem{},
		&pharmacyEntity.Doctor{},
		&pharmacyEntity.ControlledSubstanceLog{},
		&minimarketEntity.Promotion{},
		&minimarketEntity.LoyaltyProgram{},
		&minimarketEntity.CustomerLoyalty{},
		&minimarketEntity.LoyaltyTransaction{},
		&hardwareEntity.SerialNumber{},
		&hardwareEntity.WarrantyClaim{},
		&hardwareEntity.ServiceOrder{},
		&dashboardEntity.Report{},
		&dashboardEntity.ReportExecution{},
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}