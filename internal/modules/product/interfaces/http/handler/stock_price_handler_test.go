package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/testutil"
)

func (s *testServer) request(t *testing.T, method, path, token string, body any) (int, apiResponse) {
	t.Helper()
	var raw []byte
	if text, ok := body.(string); ok {
		raw = []byte(text)
	} else if body != nil {
		raw, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	var result apiResponse
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("invalid response: %s", payload)
	}
	return resp.StatusCode, result
}

func (s *testServer) seedProduct(t *testing.T) (string, map[string]any) {
	t.Helper()
	token := s.manager(t, companyA)
	categoryID := s.createCategory(t, token, "Bebidas")
	status, resp := s.post(t, "/api/v1/products", token, productBody(categoryID))
	if status != 201 {
		t.Fatalf("create product: %d %+v", status, resp.Error)
	}
	var product map[string]any
	if err := json.Unmarshal(resp.Data, &product); err != nil {
		t.Fatal(err)
	}
	return product["id"].(string), product
}

func TestGetProductIncludesStockTotal(t *testing.T) {
	s := newTestServer(t)
	id, _ := s.seedProduct(t)
	companyID, productID := uuid.MustParse(companyA), uuid.MustParse(id)
	s.stock.Balances = []testutil.StockBalance{
		{CompanyID: companyID, ProductID: productID, BranchID: 1, LocationID: 1, Quantity: decimal.RequireFromString("2.125")},
		{CompanyID: companyID, ProductID: productID, BranchID: 1, LocationID: 2, Quantity: decimal.RequireFromString("0.875")},
		{CompanyID: companyID, ProductID: productID, BranchID: 2, LocationID: 3, Quantity: decimal.RequireFromString("5")},
		{CompanyID: uuid.MustParse(companyB), ProductID: productID, Quantity: decimal.NewFromInt(100)},
		{CompanyID: companyID, ProductID: uuid.New(), Quantity: decimal.NewFromInt(100)},
	}
	token := s.token(t, companyA, "products.read")
	status, resp := s.request(t, fiber.MethodGet, "/api/v1/products/"+id, token, nil)
	var product map[string]any
	_ = json.Unmarshal(resp.Data, &product)
	if status != 200 || product["stock_total"] != "8.000" || product["id"] != id {
		t.Fatalf("stock response: %d %+v", status, product)
	}
	s.stock.Balances = nil
	status, resp = s.request(t, fiber.MethodGet, "/api/v1/products/"+id, token, nil)
	_ = json.Unmarshal(resp.Data, &product)
	if status != 200 || product["stock_total"] != "0.000" {
		t.Fatalf("empty stock: %d %+v", status, product)
	}
	status, _ = s.request(t, fiber.MethodGet, "/api/v1/products/"+id, s.token(t, companyB, "products.read"), nil)
	if status != 404 {
		t.Fatalf("foreign product status = %d", status)
	}
	status, _ = s.request(t, fiber.MethodGet, "/api/v1/products/"+id, s.token(t, companyA), nil)
	if status != 403 {
		t.Fatalf("missing read permission status = %d", status)
	}
}

func TestPatchProductPriceAndHistory(t *testing.T) {
	s := newTestServer(t)
	id, original := s.seedProduct(t)
	token := s.token(t, companyA, "products.update", "products.read")
	path := "/api/v1/products/" + id + "/price"
	status, resp := s.request(t, fiber.MethodPatch, path, token, map[string]any{
		"cost_price": 4.00, "sale_price": "6.00", "reason": " Ajuste por costo ",
	})
	if status != 200 {
		t.Fatalf("patch: %d %+v", status, resp.Error)
	}
	var updated map[string]any
	_ = json.Unmarshal(resp.Data, &updated)
	for key, value := range original {
		if key == "cost_price" || key == "sale_price" || key == "updated_at" {
			continue
		}
		if updated[key] != value {
			t.Fatalf("changed %s: %v != %v", key, updated[key], value)
		}
	}
	if updated["cost_price"] != "4.00" || updated["sale_price"] != "6.00" {
		t.Fatalf("unexpected prices: %+v", updated)
	}
	status, resp = s.request(t, fiber.MethodGet, "/api/v1/products/"+id+"/price-history", token, nil)
	var history []map[string]any
	_ = json.Unmarshal(resp.Data, &history)
	if status != 200 || len(history) != 1 || history[0]["previous_price"] != "5.50" || history[0]["new_price"] != "6.00" ||
		history[0]["previous_cost"] != "3.20" || history[0]["new_cost"] != "4.00" || history[0]["reason"] != "Ajuste por costo" {
		t.Fatalf("history: %d %+v", status, history)
	}
	status, resp = s.request(t, fiber.MethodPatch, path, token, map[string]any{"sale_price": 6.50})
	_ = json.Unmarshal(resp.Data, &updated)
	if status != 200 || updated["cost_price"] != "4.00" || updated["sale_price"] != "6.50" {
		t.Fatalf("partial patch: %d %+v", status, updated)
	}
}

func TestPatchProductPriceErrors(t *testing.T) {
	s := newTestServer(t)
	id, _ := s.seedProduct(t)
	token := s.token(t, companyA, "products.update")
	path := "/api/v1/products/" + id + "/price"
	tests := []struct {
		name, path, token string
		body              any
		status            int
	}{
		{"below cost", path, token, map[string]any{"sale_price": 2.90}, 400},
		{"negative cost", path, token, map[string]any{"cost_price": -1}, 400},
		{"precision", path, token, map[string]any{"sale_price": 6.123}, 400},
		{"empty", path, token, map[string]any{}, 400},
		{"null prices", path, token, `{"sale_price":null,"cost_price":null}`, 400},
		{"invalid decimal", path, token, `{"sale_price":"abc"}`, 400},
		{"malformed JSON", path, token, `{"sale_price":`, 400},
		{"reason length", path, token, map[string]any{"sale_price": 6, "reason": strings.Repeat("x", 161)}, 400},
		{"invalid ID", "/api/v1/products/invalid/price", token, map[string]any{"sale_price": 6}, 400},
		{"not found", "/api/v1/products/" + uuid.NewString() + "/price", token, map[string]any{"sale_price": 6}, 404},
		{"other company", path, s.token(t, companyB, "products.update"), map[string]any{"sale_price": 6}, 404},
		{"no permission", path, s.token(t, companyA), map[string]any{"sale_price": 6}, 403},
		{"no authentication", path, "", map[string]any{"sale_price": 6}, 401},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := s.request(t, fiber.MethodPatch, tt.path, tt.token, tt.body)
			if status != tt.status || resp.Success {
				t.Fatalf("status = %d, want %d; %+v", status, tt.status, resp.Error)
			}
		})
	}
	if len(s.history.Entries[uuid.MustParse(id)]) != 0 {
		t.Fatal("invalid requests created history")
	}
	if !s.products.Products[uuid.MustParse(id)].SalePrice.Equal(decimal.RequireFromString("5.50")) {
		t.Fatal("invalid requests changed price")
	}
}
