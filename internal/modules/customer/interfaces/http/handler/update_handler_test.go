package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type customerDetail struct {
	ID             string  `json:"id"`
	CompanyID      string  `json:"company_id"`
	LegalName      string  `json:"legal_name"`
	DocumentType   string  `json:"document_type"`
	DocumentNumber *string `json:"document_number"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	Address        *string `json:"address"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      *string `json:"updated_at"`
}

func (s *testServer) editor(t *testing.T, companyID string) string {
	t.Helper()
	return s.token(t, companyID, "customers.update")
}

func (s *testServer) send(t *testing.T, method, path, token string, body any) (int, apiResponse) {
	t.Helper()
	var raw []byte
	switch b := body.(type) {
	case nil:
	case string:
		raw = []byte(b)
	default:
		raw, _ = json.Marshal(b)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	var parsed apiResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("response is not JSON (%d): %s", resp.StatusCode, payload)
	}
	return resp.StatusCode, parsed
}

func (s *testServer) patch(t *testing.T, id uuid.UUID, token string, body any) (int, apiResponse, customerDetail) {
	t.Helper()
	status, resp := s.send(t, fiber.MethodPatch, "/api/v1/customers/"+id.String(), token, body)
	var customer customerDetail
	if resp.Success {
		_ = json.Unmarshal(resp.Data, &customer)
	}
	return status, resp, customer
}

// stored returns a copy of the persisted customer.
func (s *testServer) stored(t *testing.T, id uuid.UUID) entity.Customer {
	t.Helper()
	c, ok := s.customers.Get(id)
	if !ok {
		t.Fatalf("customer %s is not stored", id)
	}
	return c
}

func strOrNil(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func assertUnchanged(t *testing.T, before, after entity.Customer) {
	t.Helper()
	if before.LegalName != after.LegalName || before.DocumentType != after.DocumentType ||
		strOrNil(before.DocumentNumber) != strOrNil(after.DocumentNumber) ||
		strOrNil(before.Phone) != strOrNil(after.Phone) || strOrNil(before.Email) != strOrNil(after.Email) ||
		strOrNil(before.Address) != strOrNil(after.Address) || before.Status != after.Status ||
		(before.UpdatedAt == nil) != (after.UpdatedAt == nil) ||
		(before.UpdatedAt != nil && !before.UpdatedAt.Equal(*after.UpdatedAt)) {
		t.Errorf("customer changed:\n before %+v\n after  %+v", before, after)
	}
}

func TestPatchCustomerEndpointUpdatesSingleField(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	before := s.stored(t, customer.ID)

	status, resp, got := s.patch(t, customer.ID, s.editor(t, companyA), map[string]any{"phone": " +59171111111 "})

	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if strOrNil(got.Phone) != "+59171111111" {
		t.Errorf("phone = %s, want trimmed +59171111111", strOrNil(got.Phone))
	}
	if got.ID != customer.ID.String() || got.CompanyID != companyA || got.LegalName != "Ferreteria Central" ||
		strOrNil(got.DocumentNumber) != "1020304050" || got.DocumentType != "tax_id" ||
		strOrNil(got.Email) != strOrNil(before.Email) || got.Status != "active" {
		t.Errorf("other fields changed: %+v", got)
	}
	after := s.stored(t, customer.ID)
	if strOrNil(after.Phone) != "+59171111111" || after.LegalName != before.LegalName {
		t.Errorf("stored customer = %+v", after)
	}
}

func TestPatchCustomerEndpointUpdatesSeveralFields(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)

	status, resp, got := s.patch(t, customer.ID, s.editor(t, companyA), map[string]any{
		"legal_name":      "  Ferreteria Central SRL ",
		"document_type":   "national_id",
		"document_number": "7654321",
		"email":           "ventas@central.com",
		"address":         "Av. Blanco Galindo km 4",
	})

	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	want := customerDetail{
		LegalName:      "Ferreteria Central SRL",
		DocumentType:   "national_id",
		DocumentNumber: new("7654321"),
		Email:          new("ventas@central.com"),
		Address:        new("Av. Blanco Galindo km 4"),
	}
	if got.LegalName != want.LegalName || got.DocumentType != want.DocumentType ||
		strOrNil(got.DocumentNumber) != *want.DocumentNumber || strOrNil(got.Email) != *want.Email ||
		strOrNil(got.Address) != *want.Address {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if got.Phone == nil {
		t.Error("phone was not sent and must keep its value")
	}
}

func TestPatchCustomerEndpointClearsOptionalFields(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)

	status, resp, got := s.patch(t, customer.ID, s.editor(t, companyA), `{"email": null, "phone": "  "}`)

	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if got.Email != nil || got.Phone != nil {
		t.Errorf("email/phone should be cleared, got %+v", got)
	}
	if strOrNil(got.DocumentNumber) != "1020304050" {
		t.Errorf("document_number must be kept, got %s", strOrNil(got.DocumentNumber))
	}
}

func TestPatchCustomerEndpointAllowsOwnDocument(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	email := strOrNil(customer.Email)

	status, resp, got := s.patch(t, customer.ID, s.editor(t, companyA), map[string]any{
		"document_type":   "tax_id",
		"document_number": "1020304050",
		"email":           email,
		"legal_name":      "Ferreteria Central Renovada",
	})

	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if got.LegalName != "Ferreteria Central Renovada" {
		t.Errorf("legal_name = %s", got.LegalName)
	}
}

func TestPatchCustomerEndpointRejectsDocumentOfAnotherCustomer(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	s.seed(t, companyA, "Supermercado Norte", "7788990011", entity.CustomerStatusActive)
	before := s.stored(t, customer.ID)

	status, resp, _ := s.patch(t, customer.ID, s.editor(t, companyA), map[string]any{
		"document_number": "7788990011",
		"legal_name":      "Cambio que no debe guardarse",
	})

	if status != fiber.StatusConflict || resp.Error == nil || resp.Error.Code != "CONFLICT" {
		t.Fatalf("got %d %+v, want 409 CONFLICT", status, resp.Error)
	}
	if _, ok := resp.Error.Details["document_number"]; !ok {
		t.Errorf("409 must name the conflicting field, details = %v", resp.Error.Details)
	}
	assertUnchanged(t, before, s.stored(t, customer.ID))
}

func TestPatchCustomerEndpointAllowsDocumentUsedInAnotherCompany(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	s.seed(t, companyB, "Otra Empresa", "5555555555", entity.CustomerStatusActive)

	status, resp, got := s.patch(t, customer.ID, s.editor(t, companyA), map[string]any{"document_number": "5555555555"})

	if status != fiber.StatusOK || strOrNil(got.DocumentNumber) != "5555555555" {
		t.Fatalf("got %d %+v (error %+v), want 200", status, got, resp.Error)
	}
}

func TestPatchCustomerEndpointAllowsSameNumberWithAnotherDocumentType(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	s.seed(t, companyA, "Supermercado Norte", "7788990011", entity.CustomerStatusActive)

	status, resp, _ := s.patch(t, customer.ID, s.editor(t, companyA), map[string]any{
		"document_type":   "passport",
		"document_number": "7788990011",
	})

	if status != fiber.StatusOK {
		t.Fatalf("got %d %+v, want 200 (uniqueness is per document type)", status, resp.Error)
	}
}

func TestPatchCustomerEndpointValidatesValues(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	before := s.stored(t, customer.ID)
	token := s.editor(t, companyA)

	tests := []struct {
		name   string
		body   any
		detail string
	}{
		{name: "malformed email", body: map[string]any{"email": "no-es-un-email"}, detail: "email"},
		{name: "email too long", body: map[string]any{"email": strings.Repeat("a", 150) + "@example.com"}, detail: "email"},
		{name: "empty legal_name", body: map[string]any{"legal_name": ""}, detail: "legal_name"},
		{name: "blank legal_name", body: map[string]any{"legal_name": "   "}, detail: "legal_name"},
		{name: "null legal_name", body: `{"legal_name": null}`, detail: "legal_name"},
		{name: "legal_name too long", body: map[string]any{"legal_name": strings.Repeat("a", 161)}, detail: "legal_name"},
		{name: "unknown document_type", body: map[string]any{"document_type": "dni"}, detail: "document_type"},
		{name: "null document_type", body: `{"document_type": null}`, detail: "document_type"},
		{name: "document_number too long", body: map[string]any{"document_number": strings.Repeat("1", 31)}, detail: "document_number"},
		{name: "phone too long", body: map[string]any{"phone": strings.Repeat("7", 31)}, detail: "phone"},
		{name: "address too long", body: map[string]any{"address": strings.Repeat("a", 201)}, detail: "address"},
		{name: "number instead of string", body: `{"phone": 70000000}`, detail: "phone"},
		{name: "valid field with invalid field", body: map[string]any{"phone": "+59170000000", "email": "mal"}, detail: "email"},
	}

	t.Run("reports every invalid field at once", func(t *testing.T) {
		status, resp, _ := s.patch(t, customer.ID, token, map[string]any{"legal_name": "  ", "email": "mal", "status": "blocked"})
		if status != fiber.StatusBadRequest || resp.Error == nil {
			t.Fatalf("got %d %+v, want 400", status, resp.Error)
		}
		for _, field := range []string{"legal_name", "email", "status"} {
			if _, ok := resp.Error.Details[field]; !ok {
				t.Errorf("missing detail for %q in %v", field, resp.Error.Details)
			}
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.patch(t, customer.ID, token, tt.body)
			if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("got %d %+v, want 400 VALIDATION_ERROR", status, resp.Error)
			}
			if _, ok := resp.Error.Details[tt.detail]; !ok {
				t.Errorf("missing detail for %q in %v", tt.detail, resp.Error.Details)
			}
		})
	}

	assertUnchanged(t, before, s.stored(t, customer.ID))
}

func TestPatchCustomerEndpointRejectsInvalidBody(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	before := s.stored(t, customer.ID)
	token := s.editor(t, companyA)

	tests := []struct {
		name   string
		body   any
		code   string
		detail string
	}{
		{name: "missing body", body: nil, code: "VALIDATION_ERROR"},
		{name: "empty object", body: `{}`, code: "VALIDATION_ERROR"},
		{name: "malformed json", body: `{"legal_name":`, code: "BAD_REQUEST"},
		{name: "array body", body: `[{"legal_name":"x"}]`, code: "BAD_REQUEST"},
		{name: "null body", body: `null`, code: "BAD_REQUEST"},
		{name: "id", body: map[string]any{"id": uuid.NewString()}, code: "VALIDATION_ERROR", detail: "id"},
		{name: "customer_id", body: map[string]any{"customer_id": uuid.NewString()}, code: "VALIDATION_ERROR", detail: "customer_id"},
		{name: "company_id", body: map[string]any{"company_id": companyB}, code: "VALIDATION_ERROR", detail: "company_id"},
		{name: "status", body: map[string]any{"status": "inactive"}, code: "VALIDATION_ERROR", detail: "status"},
		{name: "credit_limit", body: map[string]any{"credit_limit": 1000}, code: "VALIDATION_ERROR", detail: "credit_limit"},
		{name: "credit_balance", body: map[string]any{"credit_balance": 1}, code: "VALIDATION_ERROR", detail: "credit_balance"},
		{name: "points_accrued", body: map[string]any{"points_accrued": 99}, code: "VALIDATION_ERROR", detail: "points_accrued"},
		{name: "created_at", body: map[string]any{"created_at": "2020-01-01T00:00:00Z"}, code: "VALIDATION_ERROR", detail: "created_at"},
		{name: "updated_at", body: map[string]any{"updated_at": "2020-01-01T00:00:00Z"}, code: "VALIDATION_ERROR", detail: "updated_at"},
		{name: "deleted_at", body: map[string]any{"deleted_at": "2020-01-01T00:00:00Z"}, code: "VALIDATION_ERROR", detail: "deleted_at"},
		{name: "unknown field", body: map[string]any{"nickname": "Tornillo"}, code: "VALIDATION_ERROR", detail: "nickname"},
		{name: "editable field with read-only field", body: map[string]any{"legal_name": "Nuevo", "status": "blocked"}, code: "VALIDATION_ERROR", detail: "status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.patch(t, customer.ID, token, tt.body)
			if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want 400 %s", status, resp.Error, tt.code)
			}
			if tt.detail != "" {
				if _, ok := resp.Error.Details[tt.detail]; !ok {
					t.Errorf("missing detail for %q in %v", tt.detail, resp.Error.Details)
				}
			}
		})
	}

	assertUnchanged(t, before, s.stored(t, customer.ID))
}

func TestPatchCustomerEndpointNotFound(t *testing.T) {
	s := newTestServer(t)
	foreign := s.seed(t, companyB, "Cliente de Empresa B", "1020304050", entity.CustomerStatusActive)
	deleted := s.seed(t, companyA, "Cliente Eliminado", "3030303030", entity.CustomerStatusActive)
	deleted.SoftDelete(clock.NewRealClock())
	_ = s.customers.Update(context.Background(), deleted)
	foreignBefore := s.stored(t, foreign.ID)
	token := s.editor(t, companyA)

	tests := []struct {
		name string
		id   uuid.UUID
	}{
		{name: "nonexistent customer", id: uuid.New()},
		{name: "customer of another company", id: foreign.ID},
		{name: "deleted customer", id: deleted.ID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.patch(t, tt.id, token, map[string]any{"legal_name": "Intento"})
			if status != fiber.StatusNotFound || resp.Error == nil || resp.Error.Code != "NOT_FOUND" ||
				resp.Error.Message != "customer not found" {
				t.Fatalf("got %d %+v, want 404 customer not found", status, resp.Error)
			}
		})
	}

	assertUnchanged(t, foreignBefore, s.stored(t, foreign.ID))
}

func TestPatchCustomerEndpointRejectsInvalidID(t *testing.T) {
	s := newTestServer(t)

	for _, id := range []string{"123", "not-a-uuid"} {
		status, resp := s.send(t, fiber.MethodPatch, "/api/v1/customers/"+id, s.editor(t, companyA), map[string]any{"legal_name": "X"})
		if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Message != "invalid customer id" {
			t.Fatalf("%s: got %d %+v, want 400 invalid customer id", id, status, resp.Error)
		}
	}
}

func TestPatchCustomerEndpointAuthorization(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	before := s.stored(t, customer.ID)
	body := map[string]any{"legal_name": "Sin permiso"}

	tests := []struct {
		name   string
		token  string
		status int
		code   string
	}{
		{name: "without token", token: "", status: 401, code: "UNAUTHORIZED"},
		{name: "invalid token", token: "not-a-jwt", status: 401, code: "UNAUTHORIZED"},
		{name: "without any permission", token: s.token(t, companyA), status: 403, code: "FORBIDDEN"},
		{name: "with only customers.read", token: s.token(t, companyA, "customers.read", "customers.list"), status: 403, code: "FORBIDDEN"},
		{name: "with only customers.status", token: s.token(t, companyA, "customers.status"), status: 403, code: "FORBIDDEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.patch(t, customer.ID, tt.token, body)
			if status != tt.status || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want %d %s", status, resp.Error, tt.status, tt.code)
			}
		})
	}

	t.Run("token without company", func(t *testing.T) {
		status, resp, _ := s.patch(t, customer.ID, s.editor(t, ""), body)
		if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Message != "company is required" {
			t.Fatalf("got %d %+v, want 400 company is required", status, resp.Error)
		}
	})

	assertUnchanged(t, before, s.stored(t, customer.ID))
}

func TestPatchCustomerEndpointDoesNotCreateCustomers(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	s.seed(t, companyA, "Supermercado Norte", "7788990011", entity.CustomerStatusActive)
	token := s.editor(t, companyA)
	count := s.customers.Count()

	bodies := []any{
		map[string]any{"legal_name": "Ferreteria Central SRL"},
		map[string]any{"document_number": "7788990011"},
		map[string]any{"email": "mal"},
	}
	for _, body := range bodies {
		s.patch(t, customer.ID, token, body)
	}
	s.patch(t, uuid.New(), token, map[string]any{"legal_name": "No existe"})

	if s.customers.Count() != count {
		t.Errorf("customer count = %d, want %d", s.customers.Count(), count)
	}
	after := s.stored(t, customer.ID)
	if after.ID != customer.ID || after.LegalName != "Ferreteria Central SRL" || !after.CreatedAt.Equal(customer.CreatedAt) {
		t.Errorf("customer must keep its id and created_at: %+v", after)
	}
}

func TestPatchCustomerEndpointTracksUpdatedAt(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	s.seed(t, companyA, "Supermercado Norte", "7788990011", entity.CustomerStatusActive)
	token := s.editor(t, companyA)
	if customer.UpdatedAt != nil {
		t.Fatalf("seeded customer should not have updated_at")
	}

	t.Run("invalid edits keep updated_at", func(t *testing.T) {
		s.patch(t, customer.ID, token, map[string]any{"email": "mal"})
		s.patch(t, customer.ID, token, map[string]any{"document_number": "7788990011"})
		if s.stored(t, customer.ID).UpdatedAt != nil {
			t.Error("updated_at changed after a rejected edit")
		}
	})

	t.Run("valid edit sets updated_at", func(t *testing.T) {
		_, _, got := s.patch(t, customer.ID, token, map[string]any{"phone": "+59172222222"})
		after := s.stored(t, customer.ID)
		want := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
		if after.UpdatedAt == nil || !after.UpdatedAt.Equal(want) || got.UpdatedAt == nil {
			t.Errorf("updated_at = %v (response %v), want %v", after.UpdatedAt, got.UpdatedAt, want)
		}
	})
}

func TestPatchCustomerEndpointEditsAnyStatus(t *testing.T) {
	s := newTestServer(t)
	token := s.editor(t, companyA)

	for _, status := range []entity.CustomerStatus{entity.CustomerStatusActive, entity.CustomerStatusInactive, entity.CustomerStatusBlocked} {
		t.Run(string(status), func(t *testing.T) {
			customer := s.seed(t, companyA, "Cliente "+string(status), "", status)
			code, resp, got := s.patch(t, customer.ID, token, map[string]any{"address": "Calle 1"})
			if code != fiber.StatusOK {
				t.Fatalf("got %d %+v, want 200", code, resp.Error)
			}
			if got.Status != string(status) {
				t.Errorf("status = %s, PATCH must not change it (want %s)", got.Status, status)
			}
		})
	}
}

func TestPatchCustomerEndpointReturnsSameShapeAsGetByID(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)

	_, patched := s.send(t, fiber.MethodPatch, "/api/v1/customers/"+customer.ID.String(), s.editor(t, companyA), map[string]any{"address": "Calle 1"})
	_, fetched := s.send(t, fiber.MethodGet, "/api/v1/customers/"+customer.ID.String(), s.token(t, companyA, "customers.read"), nil)

	if string(patched.Data) != string(fetched.Data) {
		t.Errorf("PATCH response differs from GET /:id\n patch %s\n get   %s", patched.Data, fetched.Data)
	}
}

func TestPutCustomerEndpointKeepsReplacingAllFields(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	s.seed(t, companyA, "Supermercado Norte", "7788990011", entity.CustomerStatusActive)
	token := s.editor(t, companyA)

	status, resp := s.send(t, fiber.MethodPut, "/api/v1/customers/"+customer.ID.String(), token, map[string]any{"legal_name": "Solo Nombre"})
	var got customerDetail
	_ = json.Unmarshal(resp.Data, &got)
	if status != fiber.StatusOK || got.LegalName != "Solo Nombre" || got.Phone != nil || got.DocumentNumber != nil {
		t.Fatalf("PUT must keep replacing every field: %d %+v %+v", status, got, resp.Error)
	}

	status, resp = s.send(t, fiber.MethodPut, "/api/v1/customers/"+customer.ID.String(), token, map[string]any{
		"legal_name": "Duplicado", "document_type": "tax_id", "document_number": "7788990011",
	})
	if status != fiber.StatusConflict || resp.Error.Code != "CONFLICT" {
		t.Fatalf("PUT duplicate document: got %d %+v, want 409", status, resp.Error)
	}
}

func TestPutCustomerEndpointKeepsDocumentTypeWhenOmitted(t *testing.T) {
	s := newTestServer(t)
	customer := s.seed(t, companyA, "Ferreteria Central", "1020304050", entity.CustomerStatusActive)
	token := s.editor(t, companyA)
	path := "/api/v1/customers/" + customer.ID.String()

	t.Run("omitted document_type keeps the current one", func(t *testing.T) {
		status, resp := s.send(t, fiber.MethodPut, path, token, map[string]any{
			"legal_name": "Ferreteria Central", "document_number": "1020304050",
		})
		var got customerDetail
		_ = json.Unmarshal(resp.Data, &got)
		if status != fiber.StatusOK || got.DocumentType != "tax_id" {
			t.Fatalf("got %d document_type %q (error %+v), want 200 tax_id", status, got.DocumentType, resp.Error)
		}
		if stored := s.stored(t, customer.ID); stored.DocumentType != entity.DocumentTypeTaxID {
			t.Errorf("stored document_type = %s, want tax_id", stored.DocumentType)
		}
	})

	t.Run("sent document_type is still applied", func(t *testing.T) {
		status, resp := s.send(t, fiber.MethodPut, path, token, map[string]any{
			"legal_name": "Ferreteria Central", "document_type": "passport", "document_number": "1020304050",
		})
		var got customerDetail
		_ = json.Unmarshal(resp.Data, &got)
		if status != fiber.StatusOK || got.DocumentType != "passport" {
			t.Fatalf("got %d document_type %q (error %+v), want 200 passport", status, got.DocumentType, resp.Error)
		}
	})
}
