package seed

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	companyModel "github.com/sigif/sigif-go/internal/modules/company/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/security"
)

// Seed inserts the canonical permissions, system roles, their assignments, a
// default superadmin user and a demo company with its own admin (from
// configuration) so the system is usable.
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
		adminID, err := ensureBusinessAdminRole(tx)
		if err != nil {
			return err
		}
		_, err = ensureRole(tx, entity.RoleEmployee, "Empleado sin permisos asignados", true, true)
		if err != nil {
			return err
		}

		if err := assignPermissions(tx, superadminID, allPermIDs(permIDs)); err != nil {
			return err
		}
		if err := assignPermissions(tx, adminID, []uuid.UUID{
			permIDs[entity.PermUsersList],
			permIDs[entity.PermUsersRead],
			permIDs[entity.PermPermissionsRead],
			// HU-082-02: business_admin must list the roles of its company.
			// Read-only: no create/update/delete/assign permissions are granted.
			permIDs[entity.PermRolesList],
		}); err != nil {
			return err
		}

		if err := seedDefaultAdmin(tx, cfg, superadminID); err != nil {
			return err
		}

		return seedCompanyAndAdmin(tx, cfg, superadminID, adminID)
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
		Status:      entity.RoleStatusActive.String(),
	}
	if err := tx.Create(&role).Error; err != nil {
		return uuid.Nil, err
	}
	return role.ID, nil
}

func ensureBusinessAdminRole(tx *gorm.DB) (uuid.UUID, error) {
	var role model.RoleModel
	err := tx.Where("name = ?", entity.RoleBusinessAdmin).First(&role).Error
	if err == nil {
		return role.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	// Migrate the previous role in place so existing users and permissions keep
	// their foreign-key relationships.
	err = tx.Where("name IN ?", []string{"employee", "business_admin"}).First(&role).Error
	if err == nil {
		return role.ID, tx.Model(&role).Updates(map[string]any{
			"name":        entity.RoleBusinessAdmin,
			"description": "Admin de negocio con acceso de solo lectura a usuarios",
			"is_template": true,
			"is_system":   true,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	return ensureRole(tx, entity.RoleBusinessAdmin, "Admin de negocio con acceso de solo lectura a usuarios", true, true)
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

	var existing model.UserModel
	err := tx.Where("email = ?", email).First(&existing).Error
	if err == nil {
		// Ensure the bootstrap admin is always usable: reactivate it if it was
		// deactivated or soft-deleted during testing.
		if existing.Status != entity.UserStatusActive.String() || existing.DeletedAt.Valid {
			return tx.Model(&existing).Updates(map[string]any{
				"status":     entity.UserStatusActive.String(),
				"deleted_at": nil,
			}).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
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

// seedCompanyAndAdmin creates a demo company with a superadmin user linked to
// it so business modules (product, customer) that require company_id can be
// used, plus a business_admin user that only sees the user read-only views.
func seedCompanyAndAdmin(tx *gorm.DB, cfg *config.Config, superadminID, businessAdminID uuid.UUID) error {
	company, err := ensureCompany(tx, cfg)
	if err != nil {
		return err
	}

	email := cfg.Seed.CompanyAdminEmail
	if email == "" {
		email = "admin.empresa@sigif.com"
	}

	var existing model.UserModel
	err = tx.Where("email = ?", email).First(&existing).Error
	if err == nil {
		return seedBusinessAdminUser(tx, cfg, company, businessAdminID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	password := cfg.Seed.CompanyAdminPassword
	if password == "" {
		password = "admin123"
	}
	firstName := cfg.Seed.CompanyAdminFirstName
	if firstName == "" {
		firstName = "Admin"
	}
	lastName := cfg.Seed.CompanyAdminLastName
	if lastName == "" {
		lastName = "Empresa"
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}

	companyID := company.ID
	if err := tx.Create(&model.UserModel{
		ID:                uuid.New(),
		CompanyID:         &companyID,
		RoleID:            superadminID,
		FirstName:         firstName,
		LastName:          lastName,
		Username:          email,
		Email:             email,
		PasswordHash:      hash,
		PasswordAlgorithm: security.Algorithm,
		RequiresOTP:       false,
		Status:            entity.UserStatusActive.String(),
	}).Error; err != nil {
		return err
	}

	return seedBusinessAdminUser(tx, cfg, company, businessAdminID)
}

// seedBusinessAdminUser creates the demo business_admin account so the
// read-only views can be exercised with a user that is not superadmin.
func seedBusinessAdminUser(tx *gorm.DB, cfg *config.Config, company *companyModel.CompanyModel, businessAdminID uuid.UUID) error {
	email := cfg.Seed.BusinessAdminEmail
	if email == "" {
		email = "negocio@sigif.com"
	}

	var existing model.UserModel
	err := tx.Where("email = ?", email).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	password := cfg.Seed.BusinessAdminPassword
	if password == "" {
		password = "negocio123"
	}
	firstName := cfg.Seed.BusinessAdminFirstName
	if firstName == "" {
		firstName = "Usuario"
	}
	lastName := cfg.Seed.BusinessAdminLastName
	if lastName == "" {
		lastName = "Negocio"
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}

	companyID := company.ID
	return tx.Create(&model.UserModel{
		ID:                uuid.New(),
		CompanyID:         &companyID,
		RoleID:            businessAdminID,
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

func ensureCompany(tx *gorm.DB, cfg *config.Config) (*companyModel.CompanyModel, error) {
	var company companyModel.CompanyModel
	err := tx.First(&company).Error
	if err == nil {
		return &company, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	legalName := cfg.Seed.CompanyLegalName
	if legalName == "" {
		legalName = "Empresa Demo SIGIF"
	}
	tradeName := cfg.Seed.CompanyTradeName
	if tradeName == "" {
		tradeName = "SIGIF Demo"
	}
	taxID := cfg.Seed.CompanyTaxID
	if taxID == "" {
		taxID = "100010001"
	}

	company = companyModel.CompanyModel{
		ID:        uuid.New(),
		LegalName: legalName,
		TradeName: tradeName,
		TaxID:     taxID,
		Currency:  "BOB",
		Timezone:  "America/La_Paz",
		Status:    "active",
	}
	if err := tx.Create(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}
