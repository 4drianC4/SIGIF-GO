package seed

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/model"
)

// demoRegisters are the cash registers of the demo company. There is no
// endpoint to register cash registers yet; CAJA-03 is inactive so the
// "inactive cash register" rule can be exercised.
var demoRegisters = []struct {
	code   string
	name   string
	status entity.CashRegisterStatus
}{
	{"CAJA-01", "Caja 1", entity.CashRegisterStatusActive},
	{"CAJA-02", "Caja 2", entity.CashRegisterStatusActive},
	{"CAJA-03", "Caja 3", entity.CashRegisterStatusInactive},
}

// Seed creates the demo cash registers of the company. It is idempotent: a
// code that already exists, even soft-deleted, is left untouched.
func Seed(ctx context.Context, db *gorm.DB, companyID uuid.UUID) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, r := range demoRegisters {
			var existing model.CashRegisterModel
			err := tx.Unscoped().Where("company_id = ? AND code = ?", companyID, r.code).First(&existing).Error
			if err == nil {
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Create(&model.CashRegisterModel{
				ID:        uuid.New(),
				CompanyID: companyID,
				Code:      r.code,
				Name:      r.name,
				Status:    string(r.status),
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
