package seed

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/security"
)

// Seed inserts the canonical permissions, system roles, their assignments and
// a default superadmin user (from configuration) so the system is usable.
func Seed(ctx context.Context, db *gorm.DB, cfg *config.Config) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		permIDs, err := seedPermissions(tx)
		if err != nil {
			return err
		}

		superadminID, err := ensureRole(tx, entity.RoleSuperadmin, "Acceso total al sistema", true, true)
		if err != nil {
			return err
		}
		soporteID, err := ensureRole(tx, entity.RoleSoporte, "Soporte con acceso de solo lectura a usuarios", true, true)
		if err != nil {
			return err
		}

		if err := assignPermissions(tx, superadminID, allPermIDs(permIDs)); err != nil {
			return err
		}
		if err := assignPermissions(tx, soporteID, []uuid.UUID{
			permIDs[entity.PermUsersList],
			permIDs[entity.PermUsersRead],
		}); err != nil {
			return err
		}

		return seedDefaultAdmin(tx, cfg, superadminID)
	})
}

func seedPermissions(tx *gorm.DB) (map[string]uuid.UUID, error) {
	ids := make(map[string]uuid.UUID)
	for _, p := range entity.AllPermissions() {
		var existing model.PermissionModel
		err := tx.Where("module = ? AND operation = ?", p.Module, p.Operation).First(&existing).Error
		if err == nil {
			ids[p.Key()] = existing.ID
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		m := model.PermissionModel{
			ID:          uuid.New(),
			Module:      p.Module,
			Operation:   p.Operation,
			Description: p.Description,
		}
		if err := tx.Create(&m).Error; err != nil {
			return nil, err
		}
		ids[p.Key()] = m.ID
	}
	return ids, nil
}

func ensureRole(tx *gorm.DB, name, description string, isTemplate, isSystem bool) (uuid.UUID, error) {
	var existing model.RoleModel
	err := tx.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	role := model.RoleModel{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		IsTemplate:  isTemplate,
		IsSystem:    isSystem,
	}
	if err := tx.Create(&role).Error; err != nil {
		return uuid.Nil, err
	}
	return role.ID, nil
}

func assignPermissions(tx *gorm.DB, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	for _, pid := range permissionIDs {
		if pid == uuid.Nil {
			continue
		}
		var count int64
		if err := tx.Model(&model.RolePermissionModel{}).
			Where("role_id = ? AND permission_id = ?", roleID, pid).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := tx.Create(&model.RolePermissionModel{RoleID: roleID, PermissionID: pid}).Error; err != nil {
			return err
		}
	}
	return nil
}

func allPermIDs(ids map[string]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		result = append(result, id)
	}
	return result
}

func seedDefaultAdmin(tx *gorm.DB, cfg *config.Config, superadminID uuid.UUID) error {
	var count int64
	if err := tx.Model(&model.UserModel{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	email := cfg.Seed.AdminEmail
	if email == "" {
		email = "admin@sigif.com"
	}
	password := cfg.Seed.AdminPassword
	if password == "" {
		password = "admin123"
	}
	firstName := cfg.Seed.AdminFirstName
	if firstName == "" {
		firstName = "Admin"
	}
	lastName := cfg.Seed.AdminLastName
	if lastName == "" {
		lastName = "SIGIF"
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}

	return tx.Create(&model.UserModel{
		ID:                uuid.New(),
		RoleID:            superadminID,
		FirstName:         firstName,
		LastName:          lastName,
		Username:          email,
		Email:             email,
		PasswordHash:      hash,
		PasswordAlgorithm: security.Algorithm,
		RequiresOTP:       false,
		Status:            entity.UserStatusActive.String(),
	}).Error
}
