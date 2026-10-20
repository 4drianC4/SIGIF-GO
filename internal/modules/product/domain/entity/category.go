package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

var maxTargetMargin = decimal.NewFromInt(100)

type Category struct {
	ID           uuid.UUID
	CompanyID    uuid.UUID
	ParentID     *uuid.UUID
	Name         string
	Description  string
	DefaultTaxID *uuid.UUID
	TargetMargin *decimal.Decimal
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type NewCategoryParams struct {
	CompanyID    uuid.UUID
	ParentID     *uuid.UUID
	Name         string
	Description  string
	DefaultTaxID *uuid.UUID
	TargetMargin *decimal.Decimal
}

func NewCategory(clock clock.Clock, params NewCategoryParams) (*Category, error) {
	details := validationDetails{}
	validateTargetMargin(details, params.TargetMargin)
	if err := details.err(); err != nil {
		return nil, err
	}

	now := clock.NowUTC()
	return &Category{
		ID:           uuid.New(),
		CompanyID:    params.CompanyID,
		ParentID:     params.ParentID,
		Name:         NormalizeName(params.Name),
		Description:  strings.TrimSpace(params.Description),
		DefaultTaxID: params.DefaultTaxID,
		TargetMargin: params.TargetMargin,
		Status:       StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (c *Category) IsActive() bool {
	return c.Status == StatusActive
}

// UpdateCategoryParams holds the fields a partial update may change. A nil
// pointer or an Optional that is not set leaves the field as it is; the parent
// of a category never changes.
type UpdateCategoryParams struct {
	Name         *string
	Description  *string
	DefaultTaxID Optional[uuid.UUID]
	TargetMargin Optional[decimal.Decimal]
}

// Update applies the fields that were sent. Clearing the default tax or the
// target margin makes the category inherit them from its parent.
func (c *Category) Update(clock clock.Clock, params UpdateCategoryParams) error {
	details := validationDetails{}
	if params.TargetMargin.Set {
		validateTargetMargin(details, params.TargetMargin.Value)
	}
	if err := details.err(); err != nil {
		return err
	}

	if params.Name != nil {
		c.Name = NormalizeName(*params.Name)
	}
	if params.Description != nil {
		c.Description = strings.TrimSpace(*params.Description)
	}
	if params.DefaultTaxID.Set {
		c.DefaultTaxID = params.DefaultTaxID.Value
	}
	if params.TargetMargin.Set {
		c.TargetMargin = params.TargetMargin.Value
	}
	c.UpdatedAt = clock.NowUTC()
	return nil
}

func validateTargetMargin(details validationDetails, margin *decimal.Decimal) {
	if margin == nil {
		return
	}
	if margin.IsNegative() || margin.GreaterThan(maxTargetMargin) {
		details["target_margin"] = "debe estar entre 0 y 100"
	} else if !hasAtMostDecimals(*margin, 2) {
		details["target_margin"] = "admite como máximo 2 decimales"
	}
}
