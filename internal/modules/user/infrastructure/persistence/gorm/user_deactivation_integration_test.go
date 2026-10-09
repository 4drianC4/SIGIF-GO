//go:build integration

package gorm_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	gormlib "gorm.io/gorm"

	companyModel "github.com/sigif/sigif-go/internal/modules/company/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	userGorm "github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/seed"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

func openUserDatabase(t *testing.T) *sharedDatabase.Database {
	t.Helper()
	dsn := os.Getenv("SIGIF_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SIGIF_TEST_DATABASE_DSN is not set")
	}

	admin, err := gormlib.Open(postgres.Open(dsn), &gormlib.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := admin.Exec(fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	db, err := gormlib.Open(postgres.Open(dsn+" search_path="+schema), &gormlib.Config{})
	if err != nil {
		t.Fatalf("open schema: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		admin.Exec(fmt.Sprintf(`DROP SCHEMA "%s" CASCADE`, schema))
		if sqlDB, err := admin.DB(); err == nil {
			sqlDB.Close()
		}
	})

	err = db.AutoMigrate(
		&companyModel.CompanyModel{},
		&model.UserModel{},
		&model.RoleModel{},
		&model.PermissionModel{},
		&model.RolePermissionModel{},
	)
	if err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return &sharedDatabase.Database{DB: db}
}

func createStoredUser(t *testing.T, db *sharedDatabase.Database, email string) *entity.AppUser {
	t.Helper()
	role := model.RoleModel{ID: uuid.New(), Name: "role-" + uuid.NewString()}
	if err := db.DB.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	user, err := entity.NewUser(clock.NewRealClock(), entity.RegisterUserParams{RoleID: role.ID, FirstName: "Juan", LastName: "Pérez", Email: email, Password: "password123"})
	if err != nil {
		t.Fatalf("new user: %v", err)
	}
	if err := userGorm.NewUserGormRepository(db).Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func TestDeactivatedUserIsKeptAndStillReadable(t *testing.T) {
	db := openUserDatabase(t)
	repo := userGorm.NewUserGormRepository(db)
	ctx := context.Background()
	realClock := clock.NewRealClock()
	user := createStoredUser(t, db, "juan@sigif.com")
	other := createStoredUser(t, db, "ana@sigif.com")

	user.Deactivate(realClock)
	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("update: %v", err)
	}

	var row struct {
		Status    string
		DeletedAt *string
	}
	if err := db.DB.Raw("SELECT status, deleted_at::text AS deleted_at FROM app_user WHERE id = ?", user.ID).Scan(&row).Error; err != nil {
		t.Fatalf("raw select: %v", err)
	}
	if row.Status != "inactive" || row.DeletedAt == nil {
		t.Fatalf("row must be kept as inactive with deleted_at: %+v", row)
	}

	byID, err := repo.GetByID(ctx, user.ID)
	if err != nil || byID == nil {
		t.Fatalf("deactivated user must be readable by id: %v %v", byID, err)
	}
	if byID.Status != entity.UserStatusInactive || byID.DeletedAt == nil || byID.IsActive() {
		t.Fatalf("unexpected state: %+v", byID)
	}
	if byID.RoleName == "" {
		t.Fatal("role name must still be attached")
	}

	byEmail, err := repo.GetByEmail(ctx, "juan@sigif.com")
	if err != nil || byEmail == nil || byEmail.ID != user.ID || byEmail.IsActive() {
		t.Fatalf("deactivated user must be readable by email as inactive: %+v %v", byEmail, err)
	}

	exists, err := repo.ExistsByEmail(ctx, "juan@sigif.com")
	if err != nil || !exists {
		t.Fatalf("email of a deactivated user stays reserved: %v %v", exists, err)
	}

	users, total, err := repo.List(ctx, 0, 20)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(users) != 2 {
		t.Fatalf("list must include deactivated users: total=%d rows=%d", total, len(users))
	}
	statuses := map[uuid.UUID]entity.UserStatus{}
	for _, u := range users {
		statuses[u.ID] = u.Status
	}
	if statuses[user.ID] != entity.UserStatusInactive || statuses[other.ID] != entity.UserStatusActive {
		t.Fatalf("unexpected statuses: %v", statuses)
	}
}

func TestDeactivatedUserCanBeEditedAndReactivated(t *testing.T) {
	db := openUserDatabase(t)
	repo := userGorm.NewUserGormRepository(db)
	ctx := context.Background()
	realClock := clock.NewRealClock()
	user := createStoredUser(t, db, "juan@sigif.com")

	user.Deactivate(realClock)
	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	stored, _ := repo.GetByID(ctx, user.ID)
	name := "Juan Carlos"
	stored.Edit(entity.EditUserParams{FirstName: &name}, realClock)
	if err := repo.Update(ctx, stored); err != nil {
		t.Fatalf("edit: %v", err)
	}
	stored, _ = repo.GetByID(ctx, user.ID)
	if stored.FirstName != "Juan Carlos" || stored.DeletedAt == nil || stored.Status != entity.UserStatusInactive {
		t.Fatalf("editing must keep the user deactivated: %+v", stored)
	}

	stored.Activate(realClock)
	if err := repo.Update(ctx, stored); err != nil {
		t.Fatalf("activate: %v", err)
	}
	stored, _ = repo.GetByID(ctx, user.ID)
	if !stored.IsActive() || stored.DeletedAt != nil {
		t.Fatalf("user must be active again: %+v", stored)
	}

	var count int64
	if err := db.DB.Raw("SELECT count(*) FROM app_user WHERE email = ?", "juan@sigif.com").Scan(&count).Error; err != nil || count != 1 {
		t.Fatalf("there must be exactly one row for the user: %d %v", count, err)
	}
}

func TestSeedDoesNotDuplicateDeactivatedAdmins(t *testing.T) {
	db := openUserDatabase(t)
	repo := userGorm.NewUserGormRepository(db)
	ctx := context.Background()
	cfg := &config.Config{}

	if err := seed.Seed(ctx, db.DB, cfg); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	for _, email := range []string{"admin@sigif.com", "admin.empresa@sigif.com"} {
		admin, err := repo.GetByEmail(ctx, email)
		if err != nil || admin == nil {
			t.Fatalf("seeded %s not found: %v", email, err)
		}
		admin.Deactivate(clock.NewRealClock())
		if err := repo.Update(ctx, admin); err != nil {
			t.Fatalf("deactivate %s: %v", email, err)
		}
	}

	if err := seed.Seed(ctx, db.DB, cfg); err != nil {
		t.Fatalf("seed after deactivating the admins: %v", err)
	}

	var count int64
	if err := db.DB.Raw("SELECT count(*) FROM app_user").Scan(&count).Error; err != nil || count != 2 {
		t.Fatalf("seed must not duplicate users: %d %v", count, err)
	}
	bootstrap, _ := repo.GetByEmail(ctx, "admin@sigif.com")
	if !bootstrap.IsActive() {
		t.Fatalf("bootstrap admin must be reactivated by the seed: %+v", bootstrap)
	}
}
