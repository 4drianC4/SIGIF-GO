package entity_test

import (
	"errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"strings"
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func TestSupplierValidationAndNormalization(t *testing.T) {
	clk := clock.NewMockClock(time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC))
	s, err := entity.NewSupplier(clk, uuid.New(), entity.Fields{
		LegalName: "  Andina S.R.L. ", TradeName: ptr(" Andina "), TaxID: ptr(" 001044556677 "),
		Email: ptr(" pedidos@andina.bo "), Contact: ptr("  "),
		LeadTimeDays: ptr(0), MinimumOrder: ptr(decimal.Zero),
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "Andina" || s.LegalName != "Andina S.R.L." || *s.TaxID != "001044556677" ||
		s.Contact != nil || s.Status != entity.StatusActive || !s.CreatedAt.Equal(clk.NowUTC()) || s.ID == uuid.Nil {
		t.Fatalf("unexpected normalized supplier: %+v", s)
	}
	s.TradeName = nil
	if s.Name() != s.LegalName {
		t.Fatal("name must fall back to legal_name")
	}

	cases := []struct {
		name, field string
		fields      entity.Fields
	}{
		{"blank name", "legal_name", entity.Fields{LegalName: " "}},
		{"long name", "legal_name", entity.Fields{LegalName: strings.Repeat("á", 161)}},
		{"long contact", "contact", entity.Fields{LegalName: "Andina", Contact: ptr(strings.Repeat("á", 121))}},
		{"tax punctuation", "tax_id", entity.Fields{LegalName: "Andina", TaxID: ptr("10-44")}},
		{"email display name", "email", entity.Fields{LegalName: "Andina", Email: ptr("Ventas <pedidos@andina.bo>")}},
		{"bad email", "email", entity.Fields{LegalName: "Andina", Email: ptr("no-email")}},
		{"negative days", "lead_time_days", entity.Fields{LegalName: "Andina", LeadTimeDays: ptr(-1)}},
		{"overflow days", "lead_time_days", entity.Fields{LegalName: "Andina", LeadTimeDays: ptr(32768)}},
		{"negative money", "minimum_order", entity.Fields{LegalName: "Andina", MinimumOrder: ptr(decimal.RequireFromString("-0.01"))}},
		{"money precision", "minimum_order", entity.Fields{LegalName: "Andina", MinimumOrder: ptr(decimal.RequireFromString("1.001"))}},
		{"money overflow", "minimum_order", entity.Fields{LegalName: "Andina", MinimumOrder: ptr(decimal.RequireFromString("10000000000"))}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := entity.NewSupplier(clk, uuid.New(), tc.fields)
			var appErr *sharedErrors.AppError
			if !errors.As(err, &appErr) || appErr.Code != sharedErrors.CodeValidation {
				t.Fatalf("expected validation error, got %v", err)
			}
			if appErr.Details.(map[string]string)[tc.field] == "" {
				t.Fatalf("expected detail for %s, got %v", tc.field, appErr.Details)
			}
		})
	}
}
