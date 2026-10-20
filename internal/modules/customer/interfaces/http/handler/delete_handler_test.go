package handler_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
)

func (s *testServer) remover(t *testing.T, companyID string) string {
	t.Helper()
	return s.token(t, companyID, "customers.delete")
}

// delete returns the status and the raw body: a successful delete has none.
func (s *testServer) delete(t *testing.T, id string, token string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/customers/"+id, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(payload)
}

// setBalance overwrites the credit balance and points of a stored customer,
// as the credit and loyalty modules will do.
func (s *testServer) setBalance(id uuid.UUID, balance string, points int) {
	stored := s.customers.Customers[id]
	stored.CreditBalance = decimal.RequireFromString(balance)
	stored.PointsAccrued = points
}

func TestDeleteCustomerEndpointSoftDeletes(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	s.seed(t, companyA, "Supermercado Norte", "7788990011", entity.CustomerStatusActive)
	token := s.token(t, companyA, "customers.delete", "customers.list", "customers.read")
	count := s.customers.Count()

	status, body := s.delete(t, customer.ID.String(), token)

	if status != fiber.StatusNoContent || body != "" {
		t.Fatalf("status = %d, body = %q, want 204 without body", status, body)
	}

	t.Run("row is kept with deleted_at", func(t *testing.T) {
		stored := s.stored(t, customer.ID)
		if !stored.IsDeleted() || s.customers.Count() != count {
			t.Errorf("deleted = %v, count = %d, want soft delete and %d rows", stored.IsDeleted(), s.customers.Count(), count)
		}
		if stored.LegalName != "Ferreteria Central" || strOrNil(stored.DocumentNumber) != "1020304050" {
			t.Errorf("customer data was lost: %+v", stored)
		}
	})
	t.Run("not served by id", func(t *testing.T) {
		if status, _ := s.get(t, customer.ID.String(), token); status != fiber.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})
	t.Run("not listed in any status", func(t *testing.T) {
		_, _, items := s.list(t, "status=all", token)
		if len(items) != 1 || items[0].LegalName != "Supermercado Norte" {
			t.Errorf("listing = %v, want only Supermercado Norte", names(items))
		}
	})
	t.Run("deleting again is not found", func(t *testing.T) {
		if status, _ := s.delete(t, customer.ID.String(), token); status != fiber.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})
}

func TestDeleteCustomerEndpointAllowsAnyStatus(t *testing.T) {
	for _, st := range []entity.CustomerStatus{entity.CustomerStatusActive, entity.CustomerStatusInactive, entity.CustomerStatusBlocked} {
		t.Run(string(st), func(t *testing.T) {
			s := newTestServer(t)
			customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", st)

			if status, body := s.delete(t, customer.ID.String(), s.remover(t, companyA)); status != fiber.StatusNoContent {
				t.Fatalf("status = %d, want 204 (%s)", status, body)
			}
		})
	}
}

func TestDeleteCustomerEndpointBlockedByDependencies(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(s *testServer, id uuid.UUID)
		message string
	}{
		{"related records", func(s *testServer, id uuid.UUID) { s.deps.Block(id) }, "related records"},
		{"credit balance", func(s *testServer, id uuid.UUID) { s.setBalance(id, "150.50", 0) }, "credit balance"},
		{"negative credit balance", func(s *testServer, id uuid.UUID) { s.setBalance(id, "-10", 0) }, "credit balance"},
		{"accrued points", func(s *testServer, id uuid.UUID) { s.setBalance(id, "0", 25) }, "accrued points"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(t)
			customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
			tt.prepare(s, customer.ID)
			before := s.stored(t, customer.ID)
			writes := s.customers.Writes
			token := s.token(t, companyA, "customers.delete", "customers.read")

			status, resp := s.send(t, fiber.MethodDelete, "/api/v1/customers/"+customer.ID.String(), token, nil)

			if status != fiber.StatusConflict || resp.Error == nil || resp.Error.Code != "CONFLICT" {
				t.Fatalf("status = %d, error = %+v, want 409 CONFLICT", status, resp.Error)
			}
			if !strings.Contains(resp.Error.Message, tt.message) || !strings.Contains(resp.Error.Message, "deactivate it instead") {
				t.Errorf("message = %q, want the reason and the suggestion to deactivate", resp.Error.Message)
			}

			after := s.stored(t, customer.ID)
			assertUnchanged(t, before, after)
			if after.IsDeleted() || s.customers.Writes != writes {
				t.Errorf("rejected delete changed the customer: deleted = %v, writes = %d, want %d", after.IsDeleted(), s.customers.Writes, writes)
			}
			if status, _ := s.get(t, customer.ID.String(), token); status != fiber.StatusOK {
				t.Errorf("GET after rejected delete = %d, want 200", status)
			}
		})
	}
}

func TestDeleteCustomerEndpointNotFound(t *testing.T) {
	s := newTestServer(t)
	other := s.seed(t, companyB, "Cliente Ajeno", "111", entity.CustomerStatusActive)
	deleted := s.seed(t, companyA, "Cliente Eliminado", "222", entity.CustomerStatusActive)
	s.softDelete(t, deleted)
	token := s.remover(t, companyA)

	tests := []struct {
		name string
		id   uuid.UUID
	}{
		{"unknown id", uuid.New()},
		{"customer of another company", other.ID},
		{"already deleted customer", deleted.ID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := s.send(t, fiber.MethodDelete, "/api/v1/customers/"+tt.id.String(), token, nil)

			if status != fiber.StatusNotFound || resp.Error == nil || resp.Error.Code != "NOT_FOUND" {
				t.Fatalf("status = %d, error = %+v, want 404 NOT_FOUND", status, resp.Error)
			}
		})
	}
	if stored := s.stored(t, other.ID); stored.IsDeleted() {
		t.Error("customer of another company was deleted")
	}
}

func TestDeleteCustomerEndpointRejectsInvalidID(t *testing.T) {
	s := newTestServer(t)

	status, resp := s.send(t, fiber.MethodDelete, "/api/v1/customers/not-a-uuid", s.remover(t, companyA), nil)

	if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != "BAD_REQUEST" {
		t.Fatalf("status = %d, error = %+v, want 400 BAD_REQUEST", status, resp.Error)
	}
}

func TestDeleteCustomerEndpointAuthorization(t *testing.T) {
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
		{"status permission is not enough", s.statusManager(t, companyA), fiber.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if status, _ := s.delete(t, customer.ID.String(), tt.token); status != tt.want {
				t.Errorf("status = %d, want %d", status, tt.want)
			}
		})
	}

	t.Run("token without company", func(t *testing.T) {
		status, resp := s.send(t, fiber.MethodDelete, "/api/v1/customers/"+customer.ID.String(), s.remover(t, ""), nil)
		if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Message != "company is required" {
			t.Errorf("status = %d, error = %+v, want 400 company is required", status, resp.Error)
		}
	})

	if stored := s.stored(t, customer.ID); stored.IsDeleted() {
		t.Error("customer was deleted by a rejected request")
	}
}

func TestDeletedCustomerDocumentCanBeRegisteredAgain(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	token := s.token(t, companyA, "customers.delete", "customers.create")
	body := map[string]any{"legal_name": "Ferreteria Nueva", "document_type": "tax_id", "document_number": "1020304050"}

	if status, _ := s.send(t, fiber.MethodPost, "/api/v1/customers/", token, body); status != fiber.StatusConflict {
		t.Fatalf("create before delete = %d, want 409", status)
	}
	if status, _ := s.delete(t, customer.ID.String(), token); status != fiber.StatusNoContent {
		t.Fatalf("delete = %d, want 204", status)
	}
	if status, resp := s.send(t, fiber.MethodPost, "/api/v1/customers/", token, body); status != fiber.StatusCreated {
		t.Fatalf("create after delete = %d, want 201 (error %+v)", status, resp.Error)
	}
}
