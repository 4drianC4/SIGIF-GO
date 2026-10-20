package seed

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	companyModel "github.com/sigif/sigif-go/internal/modules/company/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/security"
)

// seedCashierUsers creates two demo employees of the demo company that work as
// cashiers (RF-10), so the cash register assignment and its conflicts can be
// exercised with two different users. Existing accounts are left untouched.
func seedCashierUsers(tx *gorm.DB, cfg *config.Config, employeeID uuid.UUID) error {
	var company companyModel.CompanyModel
	if err := tx.First(&company).Error; err != nil {
		return err
	}

	cashiers := []struct {
		email, password, lastName string
	}{
		{orDefault(cfg.Seed.CashierAEmail, "cajero.a@sigif.com"), orDefault(cfg.Seed.CashierAPassword, "cajero123"), "A"},
		{orDefault(cfg.Seed.CashierBEmail, "cajero.b@sigif.com"), orDefault(cfg.Seed.CashierBPassword, "cajero123"), "B"},
	}

	for _, cashier := range cashiers {
		var existing model.UserModel
		err := tx.Where("email = ?", cashier.email).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		hash, err := security.HashPassword(cashier.password)
		if err != nil {
			return err
		}

		companyID := company.ID
		if err := tx.Create(&model.UserModel{
			ID:                uuid.New(),
			CompanyID:         &companyID,
			RoleID:            employeeID,
			FirstName:         "Cajero",
			LastName:          cashier.lastName,
			Username:          cashier.email,
			Email:             cashier.email,
			PasswordHash:      hash,
			PasswordAlgorithm: security.Algorithm,
			RequiresOTP:       false,
			Status:            entity.UserStatusActive.String(),
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
