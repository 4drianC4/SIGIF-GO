//go:build integration

package seed_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	gormlib "gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	companyModel "github.com/sigif/sigif-go/internal/modules/company/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/seed"
	"github.com/sigif/sigif-go/internal/shared/config"
)

func openSeedDatabase(t *testing.T) *gormlib.DB {
	t.Helper()
	dsn := os.Getenv("SIGIF_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SIGIF_TEST_DATABASE_DSN no está definido")
	}
	db, err := gormlib.Open(postgres.Open(dsn), &gormlib.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(
		&companyModel.CompanyModel{},
		&model.UserModel{},
		&model.RoleModel{},
		&model.PermissionModel{},
		&model.RolePermissionModel{},
	); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func cleanSeedTables(t *testing.T, db *gormlib.DB) {
	t.Helper()
	for _, table := range []string{"role_permission", "app_user", "role", "permission", "company"} {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			t.Fatalf("clean %s: %v", table, err)
		}
	}
}

func seedConfig() *config.Config {
	return &config.Config{
		Seed: config.SeedConfig{
			AdminEmail:             "admin@sigif.com",
			AdminPassword:          "admin123",
			AdminFirstName:         "Admin",
			AdminLastName:          "SIGIF",
			CompanyLegalName:       "Empresa Demo SIGIF",
			CompanyTradeName:       "SIGIF Demo",
			CompanyTaxID:           "100010001",
			CompanyAdminEmail:      "admin.empresa@sigif.com",
			CompanyAdminPassword:   "admin123",
			CompanyAdminFirstName:  "Admin",
			CompanyAdminLastName:   "Empresa",
			BusinessAdminEmail:     "negocio@sigif.com",
			BusinessAdminPassword:  "negocio123",
			BusinessAdminFirstName: "Usuario",
			BusinessAdminLastName:  "Negocio",
		},
	}
}

func roleIDByName(t *testing.T, db *gormlib.DB, name string) uuid.UUID {
	t.Helper()
	var role model.RoleModel
	if err := db.Where("name = ?", name).First(&role).Error; err != nil {
		t.Fatalf("role %q not seeded: %v", name, err)
	}
	return role.ID
}

func hasPermission(t *testing.T, db *gormlib.DB, roleID uuid.UUID, module, operation string) bool {
	t.Helper()
	var count int64
	err := db.Table("role_permission").
		Joins("JOIN permission ON permission.id = role_permission.permission_id").
		Where("role_permission.role_id = ? AND permission.module = ? AND permission.operation = ?", roleID, module, operation).
		Count(&count).Error
	if err != nil {
		t.Fatalf("count permission %s.%s: %v", module, operation, err)
	}
	return count > 0
}

func rolePermissionRows(t *testing.T, db *gormlib.DB) int64 {
	t.Helper()
	var count int64
	if err := db.Table("role_permission").Count(&count).Error; err != nil {
		t.Fatalf("count role_permission: %v", err)
	}
	return count
}

// TestSeedBusinessAdminCanListRoles verifies HU-082-02 corrections: the
// business_admin role is granted roles.list (read-only), superadmin keeps it,
// employee does not get it, and re-running the seed is idempotent.
func TestSeedBusinessAdminCanListRoles(t *testing.T) {
	db := openSeedDatabase(t)
	cleanSeedTables(t, db)
	t.Cleanup(func() { cleanSeedTables(t, db) })

	ctx := context.Background()
	cfg := seedConfig()

	if err := seed.Seed(ctx, db, cfg); err != nil {
		t.Fatalf("first Seed() error = %v", err)
	}

	superadminID := roleIDByName(t, db, entity.RoleSuperadmin)
	businessAdminID := roleIDByName(t, db, entity.RoleBusinessAdmin)
	employeeID := roleIDByName(t, db, entity.RoleEmployee)

	if !hasPermission(t, db, superadminID, "roles", "list") {
		t.Error("superadmin does not have roles.list")
	}
	if !hasPermission(t, db, businessAdminID, "roles", "list") {
		t.Error("business_admin does not have roles.list")
	}
	if hasPermission(t, db, employeeID, "roles", "list") {
		t.Error("employee has roles.list and must not")
	}

	t.Run("only the read operation is granted", func(t *testing.T) {
		for _, op := range []string{"create", "update", "delete"} {
			if hasPermission(t, db, businessAdminID, "roles", op) {
				t.Errorf("business_admin unexpectedly has roles.%s", op)
			}
		}
	})

	t.Run("seeding twice does not duplicate assignments", func(t *testing.T) {
		rows := rolePermissionRows(t, db)
		if err := seed.Seed(ctx, db, cfg); err != nil {
			t.Fatalf("second Seed() error = %v", err)
		}
		if again := rolePermissionRows(t, db); again != rows {
			t.Errorf("role_permission rows = %d after re-seed, want %d", again, rows)
		}

		var roleCount int64
		if err := db.Table("role").Count(&roleCount).Error; err != nil {
			t.Fatalf("count roles: %v", err)
		}
		var permissionCount int64
		if err := db.Table("permission").Count(&permissionCount).Error; err != nil {
			t.Fatalf("count permissions: %v", err)
		}
		if roleCount != int64(len(entity.AllSystemRoleNames())) {
			t.Errorf("roles after re-seed = %d, want %d", roleCount, len(entity.AllSystemRoleNames()))
		}
		if permissionCount != int64(len(entity.AllPermissions())) {
			t.Errorf("permissions after re-seed = %d, want %d", permissionCount, len(entity.AllPermissions()))
		}
	})

	t.Run("business_admin keeps its read permissions", func(t *testing.T) {
		if !hasPermission(t, db, businessAdminID, "users", "list") {
			t.Error("business_admin lost users.list")
		}
		if !hasPermission(t, db, businessAdminID, "users", "read") {
			t.Error("business_admin lost users.read")
		}
		if !hasPermission(t, db, businessAdminID, "permissions", "read") {
			t.Error("business_admin lost permissions.read")
		}
	})
}
