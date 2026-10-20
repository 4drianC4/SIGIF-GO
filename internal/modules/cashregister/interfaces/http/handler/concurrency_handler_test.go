package handler_test

import (
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
)

func TestAssignCashRegisterEndpointRejectsRegisterInUse(t *testing.T) {
	s := newTestServer(t)
	register := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	holderID, holder := s.cashier(t, companyA, "Cajero A")
	_, other := s.cashier(t, companyA, "Cajero B")
	_, _, held := s.assignRegister(t, holder, register.ID)
	writes := s.repo.Writes

	status, resp, _ := s.assignRegister(t, other, register.ID)

	if status != fiber.StatusConflict || errorCode(resp) != "CASH_REGISTER_IN_USE" {
		t.Fatalf("status = %d, error = %+v, want 409 CASH_REGISTER_IN_USE", status, resp.Error)
	}
	if resp.Error.Message != "cash register is already assigned to another cashier" {
		t.Errorf("message = %q", resp.Error.Message)
	}
	if s.repo.Writes != writes || s.repo.ActiveCount() != 1 {
		t.Errorf("writes = %d, active = %d, want no change", s.repo.Writes, s.repo.ActiveCount())
	}
	if stored, _ := s.repo.Assignment(uuid.MustParse(held.ID)); !stored.IsActive() || stored.UserID != holderID {
		t.Errorf("holder's assignment changed: %+v", stored)
	}
}

func TestAssignCashRegisterEndpointRejectsCashierWithRegister(t *testing.T) {
	s := newTestServer(t)
	first := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	second := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-02", entity.CashRegisterStatusActive)
	_, token := s.cashier(t, companyA, "Cajero A")
	s.assignRegister(t, token, first.ID)

	for name, register := range map[string]uuid.UUID{"another register": second.ID, "the same register again": first.ID} {
		t.Run(name, func(t *testing.T) {
			status, resp, _ := s.assignRegister(t, token, register)
			if status != fiber.StatusConflict || errorCode(resp) != "CASHIER_ALREADY_ASSIGNED" {
				t.Fatalf("status = %d, error = %+v, want 409 CASHIER_ALREADY_ASSIGNED", status, resp.Error)
			}
		})
	}
	if s.repo.ActiveCount() != 1 {
		t.Errorf("active assignments = %d, want 1", s.repo.ActiveCount())
	}
}

func TestAssignCashRegisterEndpointConcurrentRequests(t *testing.T) {
	s := newTestServer(t)
	register := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	const cashiers = 10
	tokens := make([]string, cashiers)
	for i := range tokens {
		_, tokens[i] = s.cashier(t, companyA, "Cajero")
	}

	statuses := make([]int, cashiers)
	var wg sync.WaitGroup
	for i := range tokens {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			statuses[i], _, _ = s.assignRegister(t, tokens[i], register.ID)
		}(i)
	}
	wg.Wait()

	created, conflicts := 0, 0
	for _, status := range statuses {
		switch status {
		case fiber.StatusCreated:
			created++
		case fiber.StatusConflict:
			conflicts++
		}
	}
	if created != 1 || conflicts != cashiers-1 {
		t.Errorf("created = %d, conflicts = %d, want 1 and %d (%v)", created, conflicts, cashiers-1, statuses)
	}
	if s.repo.ActiveCount() != 1 {
		t.Errorf("active assignments = %d, want 1", s.repo.ActiveCount())
	}
}

func TestReleaseAssignmentEndpointFreesRegister(t *testing.T) {
	s := newTestServer(t)
	register := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	holderID, holder := s.cashier(t, companyA, "Cajero A")
	otherID, other := s.cashier(t, companyA, "Cajero B")
	_, _, held := s.assignRegister(t, holder, register.ID)

	status, resp, got := s.release(t, held.ID, holder)

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if got.Status != "released" || got.ReleasedAt == nil || got.ReleasedBy == nil || *got.ReleasedBy != holderID.String() ||
		got.CashRegister.Code != "CAJA-01" || got.Cashier.ID != holderID.String() {
		t.Errorf("assignment = %+v, want released by the holder", got)
	}
	if _, ok := s.repo.Assignment(uuid.MustParse(held.ID)); !ok {
		t.Error("released assignment must be kept as history")
	}

	_, _, items := s.list(t, "", other)
	if item := byCode(items)["CAJA-01"]; item.Availability != "available" || item.AssignedTo != nil {
		t.Errorf("listed register = %+v, want available", item)
	}

	status, resp, reassigned := s.assignRegister(t, other, register.ID)
	if status != fiber.StatusCreated || reassigned.Cashier.ID != otherID.String() {
		t.Fatalf("reassign = %d (error %+v), want 201 for the other cashier", status, resp.Error)
	}
	if s.repo.ActiveCount() != 1 {
		t.Errorf("active assignments = %d, want 1", s.repo.ActiveCount())
	}
}

func TestReleaseAssignmentEndpointTwiceIsIdempotent(t *testing.T) {
	s := newTestServer(t)
	register := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	_, holder := s.cashier(t, companyA, "Cajero A")
	_, _, held := s.assignRegister(t, holder, register.ID)
	_, _, first := s.release(t, held.ID, holder)
	writes := s.repo.Writes

	status, resp, second := s.release(t, held.ID, holder)

	if status != fiber.StatusOK || second.Status != "released" {
		t.Fatalf("status = %d, assignment = %+v (error %+v), want 200 released", status, second, resp.Error)
	}
	if *second.ReleasedAt != *first.ReleasedAt || *second.ReleasedBy != *first.ReleasedBy {
		t.Errorf("second release = %+v, want the first release unchanged %+v", second, first)
	}
	if s.repo.Writes != writes {
		t.Errorf("writes = %d, want %d (no database write)", s.repo.Writes, writes)
	}
}

func TestReleaseAssignmentEndpointPermissions(t *testing.T) {
	s := newTestServer(t)
	register := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	_, holder := s.cashier(t, companyA, "Cajero A")
	_, other := s.cashier(t, companyA, "Cajero B")
	supervisorID, supervisor := s.user(t, companyA, "Supervisor", supervisorPerms...)

	_, _, held := s.assignRegister(t, holder, register.ID)

	t.Run("another cashier cannot release it", func(t *testing.T) {
		status, resp, _ := s.release(t, held.ID, other)
		if status != fiber.StatusForbidden || errorCode(resp) != "FORBIDDEN" {
			t.Fatalf("status = %d, error = %+v, want 403", status, resp.Error)
		}
		if stored, _ := s.repo.Assignment(uuid.MustParse(held.ID)); !stored.IsActive() {
			t.Error("assignment was released by another cashier")
		}
	})

	t.Run("a supervisor can release it", func(t *testing.T) {
		status, resp, got := s.release(t, held.ID, supervisor)
		if status != fiber.StatusOK || got.Status != "released" || got.ReleasedBy == nil || *got.ReleasedBy != supervisorID.String() {
			t.Fatalf("status = %d, assignment = %+v (error %+v), want released by the supervisor", status, got, resp.Error)
		}
	})

	t.Run("another cashier still cannot touch the released assignment", func(t *testing.T) {
		if status, _, _ := s.release(t, held.ID, other); status != fiber.StatusForbidden {
			t.Errorf("status = %d, want 403", status)
		}
	})
}

func TestReleaseAssignmentEndpointNotFoundAndInvalidID(t *testing.T) {
	s := newTestServer(t)
	_, token := s.cashier(t, companyA, "Cajero A")

	t.Run("unknown assignment", func(t *testing.T) {
		status, resp, _ := s.release(t, uuid.NewString(), token)
		if status != fiber.StatusNotFound || errorCode(resp) != "NOT_FOUND" || resp.Error.Message != "cash register assignment not found" {
			t.Errorf("status = %d, error = %+v, want 404", status, resp.Error)
		}
	})
	t.Run("invalid id", func(t *testing.T) {
		status, resp, _ := s.release(t, "not-a-uuid", token)
		if status != fiber.StatusBadRequest || errorCode(resp) != "BAD_REQUEST" || resp.Error.Message != "invalid assignment id" {
			t.Errorf("status = %d, error = %+v, want 400 invalid assignment id", status, resp.Error)
		}
	})
}
