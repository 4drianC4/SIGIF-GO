package seed

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/product/infrastructure/persistence/model"
)

var units = []model.UnitOfMeasureModel{
	{Name: "Unit", Abbreviation: "UNIT", SortOrder: 1},
	{Name: "Box", Abbreviation: "BOX", SortOrder: 2},
	{Name: "Pack", Abbreviation: "PACK", SortOrder: 3},
	{Name: "Kilogram", Abbreviation: "KG", SortOrder: 4},
	{Name: "Liter", Abbreviation: "L", SortOrder: 5},
	{Name: "Dozen", Abbreviation: "DOZ", SortOrder: 6},
}

var taxes = []model.TaxModel{
	{Name: "IVA general 13%", Percentage: decimal.NewFromInt(13)},
	{Name: "IVA reducido 5%", Percentage: decimal.NewFromInt(5)},
	{Name: "Exempt", Percentage: decimal.Zero},
}

func Seed(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, u := range units {
			if err := ensure(tx, &model.UnitOfMeasureModel{}, "abbreviation = ?", u.Abbreviation, func() any {
				u.ID = uuid.New()
				return &u
			}); err != nil {
				return err
			}
		}
		for _, t := range taxes {
			if err := ensure(tx, &model.TaxModel{}, "name = ?", t.Name, func() any {
				t.ID = uuid.New()
				return &t
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func ensure(tx *gorm.DB, existing any, condition string, value string, build func() any) error {
	err := tx.Where(condition, value).First(existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(build()).Error
}
