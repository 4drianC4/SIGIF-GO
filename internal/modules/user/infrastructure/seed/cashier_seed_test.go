//go:build integration

package seed_test

import (
	"context"
	"testing"

	companyModel "github.com/sigif/sigif-go/internal/modules/company/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/seed"
)

// TestSeedCashRegisterPermissionsAndCashiers verifies RF-10: employees work as
// cashiers, business_admin supervises the cash registers, and two demo
// cashiers of the demo company are seeded once.
func TestSeedCashRegisterPermissionsAndCashiers(t *testing.T) {
	db := openSeedDatabase(t)
	cleanSeedTables(t, db)
	t.Cleanup(func() { cleanSeedTables(t, db) })

	ctx := context.Background()
	cfg := seedConfig()
	cfg.Seed.CashierAEmail = "cajero.a@sigif.com"
	cfg.Seed.CashierBEmail = "cajero.b@sigif.com"

	if err := seed.Seed(ctx, db, cfg); err != nil {
		t.Fatalf("first Seed() error = %v", err)
	}

	employeeID := roleIDByName(t, db, entity.RoleEmployee)

	grants := []struct {
		role       string
		operations map[string]bool
	}{
		{entity.RoleSuperadmin, map[string]bool{"list": true, "assign": true, "release": true, "release_any": true}},
		{entity.RoleBusinessAdmin, map[string]bool{"list": true, "assign": false, "release": false, "release_any": true}},
		{entity.RoleEmployee, map[string]bool{"list": true, "assign": true, "release": true, "release_any": false}},
	}
	for _, g := range grants {
		for op, want := range g.operations {
			if got := hasPermission(t, db, roleIDByName(t, db, g.role), "cash_registers", op); got != want {
				t.Errorf("%s has cash_registers.%s = %v, want %v", g.role, op, got, want)
			}
		}
	}

	assertCashiers := func(t *testing.T) {
		t.Helper()
		var company companyModel.CompanyModel
		if err := db.First(&company).Error; err != nil {
			t.Fatalf("demo company: %v", err)
		}
		for _, email := range []string{"cajero.a@sigif.com", "cajero.b@sigif.com"} {
			var users []model.UserModel
			if err := db.Where("email = ?", email).Find(&users).Error; err != nil {
				t.Fatalf("find %s: %v", email, err)
			}
			if len(users) != 1 {
				t.Fatalf("%s seeded %d times, want once", email, len(users))
			}
			u := users[0]
			if u.RoleID != employeeID || u.CompanyID == nil || *u.CompanyID != company.ID ||
				u.Status != entity.UserStatusActive.String() || u.PasswordHash == "" {
				t.Errorf("%s = %+v, want an active employee of the demo company", email, u)
			}
		}
	}

	assertCashiers(t)

	t.Run("seeding twice keeps a single copy", func(t *testing.T) {
		rows := rolePermissionRows(t, db)
		if err := seed.Seed(ctx, db, cfg); err != nil {
			t.Fatalf("second Seed() error = %v", err)
		}
		if again := rolePermissionRows(t, db); again != rows {
			t.Errorf("role_permission rows = %d after re-seed, want %d", again, rows)
		}
		assertCashiers(t)
	})
}
