package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type identifierAvailability struct {
	Available       bool `json:"available"`
	ExistingProduct *struct {
		ID string `json:"id"`
	} `json:"existing_product"`
	Conflicts []conflictItem `json:"conflicts"`
}

func decodeIdentifierAvailability(t *testing.T, resp apiResponse) identifierAvailability {
	t.Helper()
	var data identifierAvailability
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("invalid availability response: %v; body=%s", err, resp.raw)
	}
	return data
}

func TestVerifyIdentifierAvailability(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	categoryID := s.createCategory(t, token, map[string]any{"name": "Abarrotes"})
	body := productBody(categoryID)
	body["barcode"] = "7750000000001"
	id := s.createProduct(t, token, body)

	status, resp := s.do(t, fiber.MethodGet, "/api/v1/products/verificar-identificador?sku=NEW-1", token, nil)
	available := decodeIdentifierAvailability(t, resp)
	if status != http.StatusOK || !available.Available || available.ExistingProduct != nil || len(available.Conflicts) != 0 {
		t.Fatalf("available identifier = %d %+v", status, available)
	}

	status, resp = s.do(t, fiber.MethodGet, "/api/v1/products/verificar-identificador?sku=%20aba-412%20&codigo_barras=7750000000001", token, nil)
	available = decodeIdentifierAvailability(t, resp)
	if status != http.StatusOK || available.Available || available.ExistingProduct == nil || available.ExistingProduct.ID != id || len(available.Conflicts) != 2 {
		t.Fatalf("occupied identifiers = %d %+v", status, available)
	}
	if available.Conflicts[0].Field != "sku" || available.Conflicts[1].Field != "barcode" {
		t.Errorf("conflicts = %+v", available.Conflicts)
	}
}

func TestVerifyIdentifierDoesNotExposeAnotherCompany(t *testing.T) {
	s := newTestServer(t)
	tokenA := s.admin(t, companyA)
	categoryID := s.createCategory(t, tokenA, map[string]any{"name": "Abarrotes"})
	body := productBody(categoryID)
	body["barcode"] = "7750000000001"
	s.createProduct(t, tokenA, body)

	tokenB := s.admin(t, companyB)
	status, resp := s.do(t, fiber.MethodGet, "/api/v1/products/verificar-identificador?sku=ABA-412&barcode=7750000000001", tokenB, nil)
	available := decodeIdentifierAvailability(t, resp)
	if status != http.StatusOK || !available.Available || available.ExistingProduct != nil || len(available.Conflicts) != 0 {
		t.Fatalf("other company response = %d %+v", status, available)
	}
}

func TestVerifyIdentifierValidationAndAuthorization(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, companyA, "products.create")

	for _, tt := range []struct {
		name   string
		path   string
		token  string
		status int
	}{
		{name: "missing identifiers", path: "/api/v1/products/verificar-identificador", token: token, status: http.StatusBadRequest},
		{name: "invalid barcode", path: "/api/v1/products/verificar-identificador?barcode=ABC", token: token, status: http.StatusBadRequest},
		{name: "missing token", path: "/api/v1/products/verificar-identificador?sku=NEW-1", status: http.StatusUnauthorized},
		{name: "missing permission", path: "/api/v1/products/verificar-identificador?sku=NEW-1", token: s.token(t, companyA), status: http.StatusForbidden},
	} {
		t.Run(tt.name, func(t *testing.T) {
			status, _ := s.do(t, fiber.MethodGet, tt.path, tt.token, nil)
			if status != tt.status {
				t.Fatalf("status = %d, want %d", status, tt.status)
			}
		})
	}
}

func TestVerifyIdentifierRateLimitIsPerUser(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, companyA, "products.create")
	path := "/api/v1/products/verificar-identificador?sku=NEW-1"

	for i := 0; i < 30; i++ {
		status, _ := s.do(t, fiber.MethodGet, path, token, nil)
		if status != http.StatusOK {
			t.Fatalf("request %d status = %d, want %d", i+1, status, http.StatusOK)
		}
	}
	status, _ := s.do(t, fiber.MethodGet, path, token, nil)
	if status != http.StatusTooManyRequests {
		t.Fatalf("request 31 status = %d, want %d", status, http.StatusTooManyRequests)
	}

	otherUser := s.token(t, companyA, "products.create")
	status, _ = s.do(t, fiber.MethodGet, path, otherUser, nil)
	if status != http.StatusOK {
		t.Fatalf("other user status = %d, want %d", status, http.StatusOK)
	}
}

// The identifier lookups share one per-user budget, so exhausting it through
// validate-duplicate also blocks verificar-identificador.
func TestIdentifierLookupsShareTheRateLimit(t *testing.T) {
	s := newTestServer(t)
	token := s.token(t, companyA, "products.create")

	for i := 0; i < 30; i++ {
		status, _ := s.do(t, fiber.MethodGet, "/api/v1/products/validate-duplicate?sku=NEW-1", token, nil)
		if status != http.StatusOK {
			t.Fatalf("request %d status = %d, want %d", i+1, status, http.StatusOK)
		}
	}
	status, _ := s.do(t, fiber.MethodGet, "/api/v1/products/verificar-identificador?sku=NEW-1", token, nil)
	if status != http.StatusTooManyRequests {
		t.Fatalf("shared budget status = %d, want %d", status, http.StatusTooManyRequests)
	}
}
