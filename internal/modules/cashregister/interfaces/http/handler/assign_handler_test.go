package handler_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
)

func TestListCashRegistersEndpointShowsActiveRegistersOfCompany(t *testing.T) {
	s := newTestServer(t)
	s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-02", entity.CashRegisterStatusActive)
	s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-03", entity.CashRegisterStatusInactive)
	s.repo.AddRegister(uuid.MustParse(companyB), "CAJA-B1", entity.CashRegisterStatusActive)
	_, token := s.cashier(t, companyA, "Cajero A")

	status, resp, items := s.list(t, "", token)

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if !slices.Equal(codes(items), []string{"CAJA-01", "CAJA-02"}) {
		t.Errorf("codes = %v, want active registers of company A ordered by code", codes(items))
	}
	for _, item := range items {
		if item.Availability != "available" || item.Status != "active" || item.AssignedTo != nil || item.AssignedAt != nil {
			t.Errorf("item = %+v, want an available active register without cashier", item)
		}
	}
	if resp.Meta == nil || resp.Meta.Total != 2 || resp.Meta.Page != 1 || resp.Meta.Limit != 20 {
		t.Errorf("meta = %+v, want page 1, limit 20, total 2", resp.Meta)
	}
}

func TestListCashRegistersEndpointAvailabilityFilter(t *testing.T) {
	s := newTestServer(t)
	held := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-02", entity.CashRegisterStatusActive)
	holderID, holderToken := s.cashier(t, companyA, "Cajero A")
	if status, resp, _ := s.assignRegister(t, holderToken, held.ID); status != fiber.StatusCreated {
		t.Fatalf("assign = %d (%+v)", status, resp.Error)
	}
	_, token := s.cashier(t, companyA, "Cajero B")

	tests := []struct {
		query string
		want  []string
	}{
		{"", []string{"CAJA-01", "CAJA-02"}},
		{"availability=all", []string{"CAJA-01", "CAJA-02"}},
		{"availability=available", []string{"CAJA-02"}},
		{"availability=in_use", []string{"CAJA-01"}},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			status, _, items := s.list(t, tt.query, token)
			if status != fiber.StatusOK || !slices.Equal(codes(items), tt.want) {
				t.Errorf("status = %d, codes = %v, want 200 %v", status, codes(items), tt.want)
			}
		})
	}

	t.Run("register in use carries its cashier", func(t *testing.T) {
		_, _, items := s.list(t, "availability=in_use", token)
		item := items[0]
		if item.Availability != "in_use" || item.AssignedTo == nil || item.AssignedTo.ID != holderID.String() ||
			item.AssignedTo.Name != "Cajero A" || item.AssignedAt == nil || item.AssignmentID == nil {
			t.Errorf("item = %+v, want in_use assigned to Cajero A", item)
		}
	})
}

func TestListCashRegistersEndpointRejectsInvalidParams(t *testing.T) {
	s := newTestServer(t)
	_, token := s.cashier(t, companyA, "Cajero A")

	tests := []struct {
		query string
		code  string
	}{
		{"availability=busy", "VALIDATION_ERROR"},
		{"availability=IN_USE", "VALIDATION_ERROR"},
		{"page=0", "VALIDATION_ERROR"},
		{"limit=0", "VALIDATION_ERROR"},
		{"limit=101", "VALIDATION_ERROR"},
		{"page=abc", "BAD_REQUEST"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			status, resp, _ := s.list(t, tt.query, token)
			if status != fiber.StatusBadRequest || errorCode(resp) != tt.code {
				t.Errorf("status = %d, error = %+v, want 400 %s", status, resp.Error, tt.code)
			}
		})
	}
}

func TestListCashRegistersEndpointPaginates(t *testing.T) {
	s := newTestServer(t)
	for i := 1; i <= 5; i++ {
		s.repo.AddRegister(uuid.MustParse(companyA), fmt.Sprintf("CAJA-%02d", i), entity.CashRegisterStatusActive)
	}
	_, token := s.cashier(t, companyA, "Cajero A")

	status, resp, items := s.list(t, "page=2&limit=2", token)

	if status != fiber.StatusOK || !slices.Equal(codes(items), []string{"CAJA-03", "CAJA-04"}) {
		t.Fatalf("status = %d, codes = %v, want CAJA-03 and CAJA-04", status, codes(items))
	}
	if resp.Meta.Total != 5 || resp.Meta.TotalPages != 3 {
		t.Errorf("meta = %+v, want total 5 and 3 pages", resp.Meta)
	}
}

func TestAssignCashRegisterEndpointAssignsAvailableRegister(t *testing.T) {
	s := newTestServer(t)
	register := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	cashierID, token := s.cashier(t, companyA, "Cajero A")

	status, resp, got := s.assignRegister(t, token, register.ID)

	if status != fiber.StatusCreated || !resp.Success {
		t.Fatalf("status = %d, want 201 (error %+v)", status, resp.Error)
	}
	if got.Status != "active" || got.CashRegister.ID != register.ID.String() || got.CashRegister.Code != "CAJA-01" ||
		got.Cashier.ID != cashierID.String() || got.Cashier.Name != "Cajero A" ||
		got.AssignedAt != "2026-10-14T08:00:00Z" || got.ReleasedAt != nil {
		t.Errorf("assignment = %+v", got)
	}

	stored, ok := s.repo.Assignment(uuid.MustParse(got.ID))
	if !ok || !stored.IsActive() || stored.UserID != cashierID || stored.CompanyID.String() != companyA {
		t.Errorf("stored = %+v, want an active assignment of the cashier in company A", stored)
	}

	_, _, items := s.list(t, "", token)
	if item := byCode(items)["CAJA-01"]; item.Availability != "in_use" || item.AssignedTo == nil || item.AssignedTo.ID != cashierID.String() {
		t.Errorf("listed register = %+v, want in_use by the cashier", item)
	}
}

func TestAssignCashRegisterEndpointIgnoresCashierInBody(t *testing.T) {
	s := newTestServer(t)
	register := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	cashierID, token := s.cashier(t, companyA, "Cajero A")
	otherID, _ := s.cashier(t, companyA, "Cajero B")

	status, _, got := s.assign(t, token, map[string]any{
		"cash_register_id": register.ID.String(),
		"user_id":          otherID.String(),
		"company_id":       companyB,
	})

	if status != fiber.StatusCreated || got.Cashier.ID != cashierID.String() {
		t.Fatalf("status = %d, cashier = %s, want 201 for the token's cashier %s", status, got.Cashier.ID, cashierID)
	}
	stored, _ := s.repo.Assignment(uuid.MustParse(got.ID))
	if stored.CompanyID.String() != companyA {
		t.Errorf("company = %s, want the token's company", stored.CompanyID)
	}
}

func TestCurrentAssignmentEndpoint(t *testing.T) {
	s := newTestServer(t)
	register := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	_, token := s.cashier(t, companyA, "Cajero A")

	t.Run("without assignment", func(t *testing.T) {
		status, resp, _ := s.current(t, token)
		if status != fiber.StatusNotFound || errorCode(resp) != "NO_ACTIVE_ASSIGNMENT" {
			t.Errorf("status = %d, error = %+v, want 404 NO_ACTIVE_ASSIGNMENT", status, resp.Error)
		}
	})

	_, _, assigned := s.assignRegister(t, token, register.ID)

	t.Run("with assignment", func(t *testing.T) {
		status, resp, got := s.current(t, token)
		if status != fiber.StatusOK || got.ID != assigned.ID || got.CashRegister.Code != "CAJA-01" || got.Status != "active" {
			t.Errorf("status = %d, assignment = %+v (error %+v), want 200 CAJA-01", status, got, resp.Error)
		}
	})

	t.Run("after release", func(t *testing.T) {
		s.release(t, assigned.ID, token)
		status, resp, _ := s.current(t, token)
		if status != fiber.StatusNotFound || errorCode(resp) != "NO_ACTIVE_ASSIGNMENT" {
			t.Errorf("status = %d, error = %+v, want 404 NO_ACTIVE_ASSIGNMENT", status, resp.Error)
		}
	})
}

func TestAssignCashRegisterEndpointNotFound(t *testing.T) {
	s := newTestServer(t)
	inactive := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-03", entity.CashRegisterStatusInactive)
	foreign := s.repo.AddRegister(uuid.MustParse(companyB), "CAJA-B1", entity.CashRegisterStatusActive)
	_, token := s.cashier(t, companyA, "Cajero A")

	tests := []struct {
		name string
		id   uuid.UUID
	}{
		{"unknown register", uuid.New()},
		{"register of another company", foreign.ID},
		{"inactive register", inactive.ID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.assignRegister(t, token, tt.id)
			if status != fiber.StatusNotFound || errorCode(resp) != "NOT_FOUND" || resp.Error.Message != "cash register not found" {
				t.Fatalf("status = %d, error = %+v, want 404 cash register not found", status, resp.Error)
			}
		})
	}
	if s.repo.ActiveCount() != 0 {
		t.Errorf("active assignments = %d, want 0", s.repo.ActiveCount())
	}
}

func TestAssignCashRegisterEndpointRejectsInvalidBody(t *testing.T) {
	s := newTestServer(t)
	_, token := s.cashier(t, companyA, "Cajero A")

	tests := []struct {
		name string
		body any
		code string
	}{
		{"empty body", map[string]any{}, "VALIDATION_ERROR"},
		{"empty id", map[string]any{"cash_register_id": ""}, "VALIDATION_ERROR"},
		{"id is not a uuid", map[string]any{"cash_register_id": "CAJA-01"}, "VALIDATION_ERROR"},
		{"id is not a string", map[string]any{"cash_register_id": 1}, "BAD_REQUEST"},
		{"malformed json", `{"cash_register_id":`, "BAD_REQUEST"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.assign(t, token, tt.body)
			if status != fiber.StatusBadRequest || errorCode(resp) != tt.code {
				t.Errorf("status = %d, error = %+v, want 400 %s", status, resp.Error, tt.code)
			}
		})
	}

	t.Run("validation error names the field", func(t *testing.T) {
		_, resp, _ := s.assign(t, token, map[string]any{})
		if resp.Error == nil || resp.Error.Details["cash_register_id"] == "" {
			t.Errorf("error = %+v, want details.cash_register_id", resp.Error)
		}
	})
}

func TestCashRegisterEndpointsAuthorization(t *testing.T) {
	s := newTestServer(t)
	register := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	_, cashierToken := s.cashier(t, companyA, "Cajero A")
	_, _, assigned := s.assignRegister(t, cashierToken, register.ID)
	assignmentID := assigned.ID
	_, noPerms := s.user(t, companyA, "Sin permisos")
	_, supervisor := s.user(t, companyA, "Supervisor", supervisorPerms...)
	assignBody := map[string]any{"cash_register_id": register.ID.String()}

	routes := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{"list", fiber.MethodGet, "/api/v1/cash-registers/", nil},
		{"assign", fiber.MethodPost, "/api/v1/cash-register-assignments/", assignBody},
		{"current", fiber.MethodGet, "/api/v1/cash-register-assignments/current", nil},
		{"release", fiber.MethodPost, "/api/v1/cash-register-assignments/" + assignmentID + "/release", nil},
	}
	for _, route := range routes {
		t.Run(route.name+" without token", func(t *testing.T) {
			if status, resp := s.send(t, route.method, route.path, "", route.body); status != fiber.StatusUnauthorized || errorCode(resp) != "UNAUTHORIZED" {
				t.Errorf("status = %d, want 401", status)
			}
		})
		t.Run(route.name+" invalid token", func(t *testing.T) {
			if status, _ := s.send(t, route.method, route.path, "not-a-token", route.body); status != fiber.StatusUnauthorized {
				t.Errorf("status = %d, want 401", status)
			}
		})
		t.Run(route.name+" without permission", func(t *testing.T) {
			if status, resp := s.send(t, route.method, route.path, noPerms, route.body); status != fiber.StatusForbidden || errorCode(resp) != "FORBIDDEN" {
				t.Errorf("status = %d, want 403", status)
			}
		})
	}

	t.Run("supervisor cannot take a cash register", func(t *testing.T) {
		if status, _, _ := s.assign(t, supervisor, assignBody); status != fiber.StatusForbidden {
			t.Errorf("status = %d, want 403", status)
		}
	})

	t.Run("token without company", func(t *testing.T) {
		_, token := s.cashier(t, "", "Sin empresa")
		for _, route := range routes[:3] {
			status, resp := s.send(t, route.method, route.path, token, route.body)
			if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Message != "company is required" {
				t.Errorf("%s: status = %d, error = %+v, want 400 company is required", route.name, status, resp.Error)
			}
		}
	})

	if stored, _ := s.repo.Assignment(uuid.MustParse(assignmentID)); !stored.IsActive() {
		t.Error("rejected requests released the assignment")
	}
}

func TestCashRegisterEndpointsIsolateCompanies(t *testing.T) {
	s := newTestServer(t)
	registerA := s.repo.AddRegister(uuid.MustParse(companyA), "CAJA-01", entity.CashRegisterStatusActive)
	registerB := s.repo.AddRegister(uuid.MustParse(companyB), "CAJA-01", entity.CashRegisterStatusActive)
	_, cashierA := s.cashier(t, companyA, "Cajero A")
	_, cashierB := s.cashier(t, companyB, "Cajero B")
	_, supervisorB := s.user(t, companyB, "Supervisor B", supervisorPerms...)
	_, _, assignedA := s.assignRegister(t, cashierA, registerA.ID)

	t.Run("company B lists only its registers, all available", func(t *testing.T) {
		_, _, items := s.list(t, "", cashierB)
		if len(items) != 1 || items[0].ID != registerB.ID.String() || items[0].Availability != "available" {
			t.Errorf("items = %+v, want only company B's available register", items)
		}
	})
	t.Run("company B cannot take company A's register", func(t *testing.T) {
		if status, _, _ := s.assignRegister(t, cashierB, registerA.ID); status != fiber.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})
	t.Run("company B supervisor cannot release company A's assignment", func(t *testing.T) {
		if status, _, _ := s.release(t, assignedA.ID, supervisorB); status != fiber.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})
	t.Run("company B cashier has no current assignment", func(t *testing.T) {
		if status, _, _ := s.current(t, cashierB); status != fiber.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})

	if stored, _ := s.repo.Assignment(uuid.MustParse(assignedA.ID)); !stored.IsActive() {
		t.Error("company A's assignment was released by company B")
	}
}
