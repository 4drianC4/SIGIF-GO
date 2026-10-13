package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

func (s *testServer) statusManager(t *testing.T, companyID string) string {
	t.Helper()
	return s.token(t, companyID, "customers.status")
}

func (s *testServer) changeStatus(t *testing.T, id string, token string, body any) (int, apiResponse, customerDetail) {
	t.Helper()
	status, resp := s.send(t, fiber.MethodPatch, "/api/v1/customers/"+id+"/status", token, body)
	var customer customerDetail
	if resp.Success {
		_ = json.Unmarshal(resp.Data, &customer)
	}
	return status, resp, customer
}

func TestChangeCustomerStatusEndpointDeactivates(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	before := s.stored(t, customer.ID)
	count := s.customers.Count()

	status, resp, got := s.changeStatus(t, customer.ID.String(), s.statusManager(t, companyA), map[string]any{"status": "inactive"})

	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if got.Status != "inactive" || got.ID != customer.ID.String() || got.UpdatedAt == nil {
		t.Errorf("response = %+v, want the inactive customer with updated_at", got)
	}

	after := s.stored(t, customer.ID)
	if after.Status != entity.CustomerStatusInactive || after.UpdatedAt == nil {
		t.Errorf("stored status = %q, updated_at = %v", after.Status, after.UpdatedAt)
	}
	if after.IsDeleted() || s.customers.Count() != count {
		t.Errorf("customer must be kept: deleted = %v, count = %d, want %d", after.IsDeleted(), s.customers.Count(), count)
	}
	// Only status and updated_at change.
	after.Status, after.UpdatedAt = before.Status, before.UpdatedAt
	assertUnchanged(t, before, after)
	if !after.CreditBalance.Equal(before.CreditBalance) || after.PointsAccrued != before.PointsAccrued ||
		!after.CreatedAt.Equal(before.CreatedAt) || after.CompanyID != before.CompanyID {
		t.Errorf("other fields changed:\n before %+v\n after  %+v", before, after)
	}
}

func TestChangeCustomerStatusEndpointTransitions(t *testing.T) {
	tests := []struct {
		from entity.CustomerStatus
		to   entity.CustomerStatus
	}{
		{entity.CustomerStatusInactive, entity.CustomerStatusActive},
		{entity.CustomerStatusBlocked, entity.CustomerStatusActive},
		{entity.CustomerStatusActive, entity.CustomerStatusBlocked},
		{entity.CustomerStatusBlocked, entity.CustomerStatusInactive},
	}
	for _, tt := range tests {
		t.Run(string(tt.from)+" to "+string(tt.to), func(t *testing.T) {
			s := newTestServer(t)
			customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", tt.from)

			status, _, got := s.changeStatus(t, customer.ID.String(), s.statusManager(t, companyA), map[string]any{"status": tt.to})

			if status != fiber.StatusOK || got.Status != string(tt.to) {
				t.Fatalf("status = %d, customer status = %q, want 200 and %q", status, got.Status, tt.to)
			}
			if stored := s.stored(t, customer.ID); stored.Status != tt.to {
				t.Errorf("stored status = %q, want %q", stored.Status, tt.to)
			}
		})
	}
}

func TestChangeCustomerStatusEndpointSameStatusIsIdempotent(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusInactive)
	before := s.stored(t, customer.ID)
	writes := s.customers.Writes

	status, resp, got := s.changeStatus(t, customer.ID.String(), s.statusManager(t, companyA), map[string]any{"status": "inactive"})

	if status != fiber.StatusOK || got.Status != "inactive" {
		t.Fatalf("status = %d, customer status = %q, want 200 inactive (error %+v)", status, got.Status, resp.Error)
	}
	if got.UpdatedAt != nil {
		t.Errorf("updated_at = %v, want it untouched", *got.UpdatedAt)
	}
	assertUnchanged(t, before, s.stored(t, customer.ID))
	if s.customers.Writes != writes {
		t.Errorf("writes = %d, want %d (no database write)", s.customers.Writes, writes)
	}
}

func TestChangeCustomerStatusEndpointRejectsInvalidBody(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	before := s.stored(t, customer.ID)
	token := s.statusManager(t, companyA)

	tests := []struct {
		name string
		body any
	}{
		{"unknown status", map[string]any{"status": "deleted"}},
		{"status with another case", map[string]any{"status": "Inactive"}},
		{"missing status", map[string]any{}},
		{"empty status", map[string]any{"status": ""}},
		{"null status", map[string]any{"status": nil}},
		{"status is not a string", map[string]any{"status": 1}},
		{"malformed json", `{"status":`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.changeStatus(t, customer.ID.String(), token, tt.body)

			if status != fiber.StatusBadRequest || resp.Success {
				t.Fatalf("status = %d, want 400", status)
			}
			assertUnchanged(t, before, s.stored(t, customer.ID))
		})
	}

	t.Run("validation error names the field", func(t *testing.T) {
		_, resp, _ := s.changeStatus(t, customer.ID.String(), token, map[string]any{"status": "deleted"})
		if resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" || resp.Error.Details["status"] == "" {
			t.Errorf("error = %+v, want VALIDATION_ERROR with details.status", resp.Error)
		}
	})
}

func TestChangeCustomerStatusEndpointNotFound(t *testing.T) {
	s := newTestServer(t)
	other := s.seed(t, companyB, "Cliente Ajeno", "111", entity.CustomerStatusActive)
	deleted := s.seed(t, companyA, "Cliente Eliminado", "222", entity.CustomerStatusActive)
	s.softDelete(t, deleted)
	token := s.statusManager(t, companyA)

	tests := []struct {
		name string
		id   uuid.UUID
	}{
		{"unknown id", uuid.New()},
		{"customer of another company", other.ID},
		{"deleted customer", deleted.ID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.changeStatus(t, tt.id.String(), token, map[string]any{"status": "inactive"})

			if status != fiber.StatusNotFound || resp.Error == nil || resp.Error.Code != "NOT_FOUND" {
				t.Fatalf("status = %d, error = %+v, want 404 NOT_FOUND", status, resp.Error)
			}
		})
	}
	if stored := s.stored(t, other.ID); stored.Status != entity.CustomerStatusActive {
		t.Errorf("customer of another company changed to %q", stored.Status)
	}
}

func TestChangeCustomerStatusEndpointRejectsInvalidID(t *testing.T) {
	s := newTestServer(t)

	status, resp, _ := s.changeStatus(t, "not-a-uuid", s.statusManager(t, companyA), map[string]any{"status": "inactive"})

	if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != "BAD_REQUEST" {
		t.Fatalf("status = %d, error = %+v, want 400 BAD_REQUEST", status, resp.Error)
	}
}

func TestChangeCustomerStatusEndpointAuthorization(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	body := map[string]any{"status": "inactive"}

	tests := []struct {
		name  string
		token string
		want  int
	}{
		{"without token", "", fiber.StatusUnauthorized},
		{"invalid token", "not-a-token", fiber.StatusUnauthorized},
		{"without permission", s.token(t, companyA), fiber.StatusForbidden},
		{"update permission is not enough", s.editor(t, companyA), fiber.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if status, _, _ := s.changeStatus(t, customer.ID.String(), tt.token, body); status != tt.want {
				t.Errorf("status = %d, want %d", status, tt.want)
			}
		})
	}

	t.Run("token without company", func(t *testing.T) {
		status, resp, _ := s.changeStatus(t, customer.ID.String(), s.statusManager(t, ""), body)
		if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Message != "company is required" {
			t.Errorf("status = %d, error = %+v, want 400 company is required", status, resp.Error)
		}
	})

	if stored := s.stored(t, customer.ID); stored.Status != entity.CustomerStatusActive {
		t.Errorf("status = %q after rejected requests, want active", stored.Status)
	}
}

func TestDeactivatedCustomerLeavesDefaultListing(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	s.seed(t, companyA, "Supermercado Norte", "7788990011", entity.CustomerStatusActive)
	token := s.token(t, companyA, "customers.status", "customers.list", "customers.read")

	if status, _, _ := s.changeStatus(t, customer.ID.String(), token, map[string]any{"status": "inactive"}); status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}

	t.Run("not listed by default", func(t *testing.T) {
		_, _, items := s.list(t, "", token)
		if len(items) != 1 || items[0].LegalName != "Supermercado Norte" {
			t.Errorf("default listing = %v, want only Supermercado Norte", names(items))
		}
	})
	t.Run("listed with the inactive filter", func(t *testing.T) {
		_, _, items := s.list(t, "status=inactive", token)
		if len(items) != 1 || items[0].ID != customer.ID.String() {
			t.Errorf("inactive listing = %v, want Ferreteria Central", names(items))
		}
	})
	t.Run("still served by id", func(t *testing.T) {
		status, resp := s.get(t, customer.ID.String(), token)
		var got customerDetail
		_ = json.Unmarshal(resp.Data, &got)
		if status != fiber.StatusOK || got.Status != "inactive" {
			t.Errorf("status = %d, customer status = %q, want 200 inactive", status, got.Status)
		}
	})
}
