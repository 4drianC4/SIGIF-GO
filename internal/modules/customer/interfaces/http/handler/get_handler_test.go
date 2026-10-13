package handler_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

func (s *testServer) viewer(t *testing.T, companyID string) string {
	t.Helper()
	return s.token(t, companyID, "customers.read")
}

func (s *testServer) get(t *testing.T, id string, token string) (int, apiResponse) {
	t.Helper()
	return s.send(t, fiber.MethodGet, "/api/v1/customers/"+id, token, nil)
}

// softDelete marks a stored customer as deleted, as DELETE /customers/:id does.
func (s *testServer) softDelete(t *testing.T, customer *entity.Customer) {
	t.Helper()
	customer.SoftDelete(clock.NewRealClock())
	if err := s.customers.Update(context.Background(), customer); err != nil {
		t.Fatalf("soft delete customer: %v", err)
	}
}

func TestGetCustomerEndpointReturnsFullCustomer(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusBlocked)

	status, resp := s.get(t, customer.ID.String(), s.viewer(t, companyA))

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("data is not an object: %s", resp.Data)
	}
	want := map[string]any{
		"id":              customer.ID.String(),
		"company_id":      companyA,
		"legal_name":      "Ferreteria Central",
		"document_type":   "tax_id",
		"document_number": "1020304050",
		"phone":           *customer.Phone,
		"email":           *customer.Email,
		"credit_limit":    "0",
		"credit_balance":  "0",
		"points_accrued":  float64(0),
		"status":          "blocked",
		"created_at":      "2026-10-01T08:01:00Z",
	}
	for field, value := range want {
		if data[field] != value {
			t.Errorf("%s = %v, want %v", field, data[field], value)
		}
	}
	if _, ok := data["deleted_at"]; ok {
		t.Errorf("deleted_at must not be exposed: %s", resp.Data)
	}
}

func TestGetCustomerEndpointReturnsEveryStatus(t *testing.T) {
	s := newTestServer(t)
	token := s.viewer(t, companyA)

	for _, st := range []entity.CustomerStatus{entity.CustomerStatusActive, entity.CustomerStatusInactive, entity.CustomerStatusBlocked} {
		t.Run(string(st), func(t *testing.T) {
			customer := s.seed(t, companyA, "Cliente "+string(st), "", st)

			status, resp := s.get(t, customer.ID.String(), token)

			var got customerDetail
			_ = json.Unmarshal(resp.Data, &got)
			if status != fiber.StatusOK || got.Status != string(st) {
				t.Errorf("status = %d, customer status = %q, want 200 and %q", status, got.Status, st)
			}
		})
	}
}

func TestGetCustomerEndpointRejectsInvalidID(t *testing.T) {
	s := newTestServer(t)

	status, resp := s.get(t, "not-a-uuid", s.viewer(t, companyA))

	if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != "BAD_REQUEST" {
		t.Fatalf("status = %d, error = %+v, want 400 BAD_REQUEST", status, resp.Error)
	}
}

func TestGetCustomerEndpointNotFound(t *testing.T) {
	s := newTestServer(t)
	other := s.seed(t, companyB, "Cliente Ajeno", "111", entity.CustomerStatusActive)
	deleted := s.seed(t, companyA, "Cliente Eliminado", "222", entity.CustomerStatusActive)
	s.softDelete(t, deleted)
	token := s.viewer(t, companyA)

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
			status, resp := s.get(t, tt.id.String(), token)

			if status != fiber.StatusNotFound || resp.Error == nil || resp.Error.Code != "NOT_FOUND" {
				t.Fatalf("status = %d, error = %+v, want 404 NOT_FOUND", status, resp.Error)
			}
			if resp.Error.Message != "customer not found" {
				t.Errorf("message = %q, must not reveal why the customer is missing", resp.Error.Message)
			}
		})
	}
}

func TestGetCustomerEndpointAuthorization(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)

	tests := []struct {
		name  string
		token string
		want  int
	}{
		{"without token", "", fiber.StatusUnauthorized},
		{"invalid token", "not-a-token", fiber.StatusUnauthorized},
		{"without permission", s.token(t, companyA), fiber.StatusForbidden},
		{"list permission is not enough", s.reader(t, companyA), fiber.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if status, _ := s.get(t, customer.ID.String(), tt.token); status != tt.want {
				t.Errorf("status = %d, want %d", status, tt.want)
			}
		})
	}

	t.Run("token without company", func(t *testing.T) {
		status, resp := s.get(t, customer.ID.String(), s.viewer(t, ""))
		if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Message != "company is required" {
			t.Errorf("status = %d, error = %+v, want 400 company is required", status, resp.Error)
		}
	})
}

func TestGetCustomerEndpointDoesNotModifyData(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	before := s.stored(t, customer.ID)
	writes := s.customers.Writes

	s.get(t, customer.ID.String(), s.viewer(t, companyA))

	assertUnchanged(t, before, s.stored(t, customer.ID))
	if s.customers.Writes != writes {
		t.Errorf("writes = %d, want %d", s.customers.Writes, writes)
	}
}
