//go:build integration

package gorm_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	gormlib "gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	userGorm "github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/gorm"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

func openHistoryDatabase(t *testing.T) *sharedDatabase.Database {
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

	if err := db.AutoMigrate(&model.UserModel{}, &model.RoleModel{}, &model.UserHistoryModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return &sharedDatabase.Database{DB: db}
}

func historyText(value string) *string { return &value }

func TestUserHistoryRepositoryRoundTrip(t *testing.T) {
	db := openHistoryDatabase(t)
	repo := userGorm.NewUserHistoryGormRepository(db)
	ctx := context.Background()
	mockClock := clock.NewMockClock(time.Date(2026, 10, 8, 18, 20, 0, 0, time.UTC))
	userID, otherID := uuid.New(), uuid.New()
	actor := &entity.HistoryActor{ID: uuid.New(), FullName: "Admin SIGIF", Email: "admin@sigif.com"}

	created := entity.NewUserHistory(mockClock, userID, entity.HistoryActionCreated, nil, nil)
	mockClock.Add(time.Minute)
	updated := entity.NewUserHistory(mockClock, userID, entity.HistoryActionUpdated, entity.HistoryChanges{
		"email":      {From: historyText("juan@sigif.com"), To: historyText("juan.carlos@sigif.com")},
		"company_id": {From: nil, To: historyText("11111111-1111-4111-8111-111111111111")},
	}, actor)
	mockClock.Add(time.Minute)
	deactivated := entity.NewUserHistory(mockClock, userID, entity.HistoryActionDeactivated, nil, actor)
	foreign := entity.NewUserHistory(mockClock, otherID, entity.HistoryActionCreated, nil, actor)

	for _, record := range []*entity.UserHistory{created, updated, deactivated, foreign} {
		if err := repo.Create(ctx, record); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	rows, total, err := repo.ListByUserID(ctx, userID, 0, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 3 || len(rows) != 2 {
		t.Fatalf("want total 3 and 2 rows, got %d and %d", total, len(rows))
	}
	if rows[0].ID != deactivated.ID || rows[1].ID != updated.ID {
		t.Fatalf("unexpected order: %s, %s", rows[0].Action, rows[1].Action)
	}

	got := rows[1]
	if got.Action != entity.HistoryActionUpdated || got.Description != "User updated" || !got.CreatedAt.Equal(updated.CreatedAt) {
		t.Fatalf("unexpected record: %+v", got)
	}
	if got.PerformedBy == nil || *got.PerformedBy != *actor {
		t.Fatalf("unexpected actor: %+v", got.PerformedBy)
	}
	email := got.Changes["email"]
	if email.From == nil || *email.From != "juan@sigif.com" || email.To == nil || *email.To != "juan.carlos@sigif.com" {
		t.Fatalf("unexpected email change: %+v", email)
	}
	company := got.Changes["company_id"]
	if company.From != nil || company.To == nil || *company.To != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("unexpected company change: %+v", company)
	}

	rest, _, err := repo.ListByUserID(ctx, userID, 2, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rest) != 1 || rest[0].ID != created.ID || rest[0].PerformedBy != nil || len(rest[0].Changes) != 0 {
		t.Fatalf("unexpected last page: %+v", rest)
	}

	empty, total, err := repo.ListByUserID(ctx, uuid.New(), 0, 20)
	if err != nil || total != 0 || len(empty) != 0 {
		t.Fatalf("unknown user must have no history: %v %d %d", err, total, len(empty))
	}
}

func TestUserHistorySurvivesUserSoftDelete(t *testing.T) {
	db := openHistoryDatabase(t)
	historyRepo := userGorm.NewUserHistoryGormRepository(db)
	ctx := context.Background()
	realClock := clock.NewRealClock()

	user := model.UserModel{ID: uuid.New(), RoleID: uuid.New(), FirstName: "Juan", LastName: "Pérez", Username: "juan@sigif.com", Email: "juan@sigif.com", PasswordHash: "hash", Status: "active"}
	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := historyRepo.Create(ctx, entity.NewUserHistory(realClock, user.ID, entity.HistoryActionCreated, nil, nil)); err != nil {
		t.Fatalf("create history: %v", err)
	}

	if err := db.DB.Delete(&user).Error; err != nil {
		t.Fatalf("soft delete user: %v", err)
	}
	if _, total, err := historyRepo.ListByUserID(ctx, user.ID, 0, 20); err != nil || total != 1 {
		t.Fatalf("history must survive a soft delete: %v %d", err, total)
	}

	if err := db.DB.Unscoped().Delete(&user).Error; err != nil {
		t.Fatalf("hard delete user: %v", err)
	}
	if _, total, err := historyRepo.ListByUserID(ctx, user.ID, 0, 20); err != nil || total != 1 {
		t.Fatalf("history must survive a hard delete: %v %d", err, total)
	}
}

func TestTransactorRollsBackUserChangeWhenHistoryFails(t *testing.T) {
	db := openHistoryDatabase(t)
	users := userGorm.NewUserGormRepository(db)
	historyRepo := userGorm.NewUserHistoryGormRepository(db)
	tx := userGorm.NewTransactor(db)
	ctx := context.Background()
	realClock := clock.NewRealClock()

	role := model.RoleModel{ID: uuid.New(), Name: "employee"}
	if err := db.DB.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	user, err := entity.NewUser(realClock, entity.RegisterUserParams{RoleID: role.ID, FirstName: "Juan", LastName: "Pérez", Email: "juan@sigif.com", Password: "password123"})
	if err != nil {
		t.Fatalf("new user: %v", err)
	}
	if err := users.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	failure := errors.New("history unavailable")
	err = tx.WithinTransaction(ctx, func(ctx context.Context) error {
		user.FirstName = "Juan Carlos"
		if err := users.Update(ctx, user); err != nil {
			return err
		}
		if err := historyRepo.Create(ctx, entity.NewUserHistory(realClock, user.ID, entity.HistoryActionUpdated, nil, nil)); err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("want the callback error, got %v", err)
	}

	stored, err := users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if stored.FirstName != "Juan" {
		t.Fatalf("user change was not rolled back: %q", stored.FirstName)
	}
	if _, total, _ := historyRepo.ListByUserID(ctx, user.ID, 0, 20); total != 0 {
		t.Fatalf("history was not rolled back: %d", total)
	}

	err = tx.WithinTransaction(ctx, func(ctx context.Context) error {
		user.FirstName = "Juan Carlos"
		if err := users.Update(ctx, user); err != nil {
			return err
		}
		return historyRepo.Create(ctx, entity.NewUserHistory(realClock, user.ID, entity.HistoryActionUpdated, nil, nil))
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	stored, _ = users.GetByID(ctx, user.ID)
	if _, total, _ := historyRepo.ListByUserID(ctx, user.ID, 0, 20); stored.FirstName != "Juan Carlos" || total != 1 {
		t.Fatalf("committed transaction was not persisted: %q %d", stored.FirstName, total)
	}
}
