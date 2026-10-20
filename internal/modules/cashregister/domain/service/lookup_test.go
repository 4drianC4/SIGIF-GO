package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
)

func TestGetRegister(t *testing.T) {
	ctx := context.Background()
	svc, repo := newService()
	companyA, companyB := uuid.New(), uuid.New()
	active := repo.AddRegister(companyA, "CAJA-01", entity.CashRegisterStatusActive)
	inactive := repo.AddRegister(companyA, "CAJA-02", entity.CashRegisterStatusInactive)

	for _, want := range []*entity.CashRegister{active, inactive} {
		got, err := svc.GetRegister(ctx, companyA, want.ID)
		if err != nil || got.ID != want.ID || got.Status != want.Status {
			t.Errorf("GetRegister(%s) = %+v, %v; want status %s", want.Code, got, err, want.Status)
		}
	}
	if _, err := svc.GetRegister(ctx, companyB, active.ID); !errors.Is(err, service.ErrCashRegisterNotFound) {
		t.Errorf("another company: error = %v, want ErrCashRegisterNotFound", err)
	}
	if _, err := svc.GetRegister(ctx, companyA, uuid.New()); !errors.Is(err, service.ErrCashRegisterNotFound) {
		t.Errorf("unknown register: error = %v, want ErrCashRegisterNotFound", err)
	}
}
