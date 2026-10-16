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
	if m := params.TargetMargin; m != nil {
		if m.IsNegative() || m.GreaterThan(maxTargetMargin) {
			details["target_margin"] = "debe estar entre 0 y 100"
		} else if !hasAtMostDecimals(*m, 2) {
			details["target_margin"] = "admite como máximo 2 decimales"
		}
	}
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
