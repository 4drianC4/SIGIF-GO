package handler_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/product/testutil"
)

// seedProduct creates a product (cost 4.20, price 6.00) and returns its id.
func (s *testServer) seedProduct(t *testing.T) string {
	t.Helper()
	admin := s.admin(t, companyA)
	categoryID := s.createCategory(t, admin, map[string]any{"name": "Bebidas"})
	return s.createProduct(t, admin, productBody(categoryID))
}

func (s *testServer) priceHistory(t *testing.T, token, id string, query ...string) (int, []map[string]any, apiResponse) {
	t.Helper()
	path := "/api/v1/products/" + id + "/price-history"
	if len(query) > 0 {
		path += "?" + query[0]
	}
	status, resp := s.do(t, fiber.MethodGet, path, token, nil)
	var entries []map[string]any
	if status == fiber.StatusOK {
		if err := json.Unmarshal(resp.Data, &entries); err != nil {
			t.Fatalf("history is not a list: %s", resp.raw)
		}
	}
	return status, entries, resp
}

func TestGetProductIncludesStockTotal(t *testing.T) {
	s := newTestServer(t)
	id := s.seedProduct(t)
	companyID, productID := uuid.MustParse(companyA), uuid.MustParse(id)
	s.store.Stock.Balances = []testutil.StockBalance{
		{CompanyID: companyID, ProductID: productID, BranchID: 1, LocationID: 1, Quantity: decimal.RequireFromString("2.125")},
		{CompanyID: companyID, ProductID: productID, BranchID: 1, LocationID: 2, Quantity: decimal.RequireFromString("0.875")},
		{CompanyID: companyID, ProductID: productID, BranchID: 2, LocationID: 3, Quantity: decimal.RequireFromString("5")},
		{CompanyID: uuid.MustParse(companyB), ProductID: productID, Quantity: decimal.NewFromInt(100)},
		{CompanyID: companyID, ProductID: uuid.New(), Quantity: decimal.NewFromInt(100)},
	}
	token := s.token(t, companyA, "products.read")

	status, resp := s.do(t, fiber.MethodGet, "/api/v1/products/"+id, token, nil)
	product := decode(t, resp.Data)
	if status != 200 || product["stock_total"] != "8.000" || product["id"] != id {
		t.Fatalf("stock response: %d %s", status, resp.raw)
	}
	// "stock" is the quantity kept on the product itself; "stock_total" is the sum
	// of the stock_balance cache. They are independent until inventory movements
	// keep both in sync.
	if !strings.Contains(resp.raw, `"stock":"0"`) {
		t.Errorf("product stock should still come from the product row: %s", resp.raw)
	}

	s.store.Stock.Balances = nil
	status, resp = s.do(t, fiber.MethodGet, "/api/v1/products/"+id, token, nil)
	if status != 200 || decode(t, resp.Data)["stock_total"] != "0.000" {
		t.Fatalf("empty stock: %d %s", status, resp.raw)
	}

	status, resp = s.do(t, fiber.MethodGet, "/api/v1/products/"+id, s.token(t, companyB, "products.read"), nil)
	expectError(t, status, resp, 404, "NOT_FOUND")
	status, resp = s.do(t, fiber.MethodGet, "/api/v1/products/"+id, s.token(t, companyA), nil)
	expectError(t, status, resp, 403, "FORBIDDEN")
}

func TestListProductsDoesNotIncludeStockTotal(t *testing.T) {
	s := newTestServer(t)
	s.seedProduct(t)
	status, resp := s.do(t, fiber.MethodGet, "/api/v1/products", s.token(t, companyA, "products.list"), nil)
	if status != 200 || strings.Contains(resp.raw, "stock_total") {
		t.Fatalf("list must keep its contract: %d %s", status, resp.raw)
	}
}

func TestPatchProductPriceAndHistory(t *testing.T) {
	s := newTestServer(t)
	id := s.seedProduct(t)
	token := s.token(t, companyA, "products.update", "products.read")
	_, resp := s.do(t, fiber.MethodGet, "/api/v1/products/"+id, token, nil)
	original := decode(t, resp.Data)

	path := "/api/v1/products/" + id + "/price"
	status, resp := s.do(t, fiber.MethodPatch, path, token, map[string]any{
		"cost": 4.00, "sale_price": "7.00", "reason": " Ajuste por costo ",
	})
	if status != 200 {
		t.Fatalf("patch: %d %s", status, resp.raw)
	}
	updated := decode(t, resp.Data)
	for _, fragment := range []string{`"cost":"4.00"`, `"price":"7.00"`} {
		if !strings.Contains(resp.raw, fragment) {
			t.Errorf("response missing %s: %s", fragment, resp.raw)
		}
	}
	for key, value := range original {
		switch key {
		case "cost", "price", "margin", "updated_at", "stock_total":
			continue
		}
		if !reflect.DeepEqual(updated[key], value) {
			t.Errorf("price patch changed %s: %v != %v", key, updated[key], value)
		}
	}

	status, history, resp := s.priceHistory(t, token, id)
	if status != 200 || len(history) != 1 || history[0]["previous_price"] != "6.00" || history[0]["new_price"] != "7.00" ||
		history[0]["previous_cost"] != "4.20" || history[0]["new_cost"] != "4.00" || history[0]["reason"] != "Ajuste por costo" {
		t.Fatalf("history: %d %s", status, resp.raw)
	}

	status, resp = s.do(t, fiber.MethodPatch, path, token, map[string]any{"sale_price": 7.50})
	if status != 200 || !strings.Contains(resp.raw, `"cost":"4.00"`) || !strings.Contains(resp.raw, `"price":"7.50"`) {
		t.Fatalf("partial patch: %d %s", status, resp.raw)
	}

	status, page, resp := s.priceHistory(t, token, id, "limit=1&page=1")
	if status != 200 || len(page) != 1 || !strings.Contains(string(resp.Meta), `"total":2`) {
		t.Fatalf("paginated history: %d %s", status, resp.raw)
	}
	if page[0]["new_price"] != "7.50" {
		t.Errorf("history must list the newest change first: %v", page[0])
	}
}

func TestPutProductRecordsPriceHistory(t *testing.T) {
	s := newTestServer(t)
	admin := s.admin(t, companyA)
	categoryID := s.createCategory(t, admin, map[string]any{"name": "Abarrotes"})
	id := s.createProduct(t, admin, productBody(categoryID))
	token := s.token(t, companyA, "products.update", "products.read")

	renamed := updateBody(categoryID)
	renamed["cost"], renamed["sale_price"] = 4.20, 6.00
	if status, resp := s.do(t, fiber.MethodPut, "/api/v1/products/"+id, token, renamed); status != 200 {
		t.Fatalf("rename: %d %s", status, resp.raw)
	}
	if _, history, _ := s.priceHistory(t, token, id); len(history) != 0 {
		t.Fatalf("a PUT that keeps the prices must not write history: %v", history)
	}

	repriced := updateBody(categoryID)
	repriced["reason"] = "Promoción"
	if status, resp := s.do(t, fiber.MethodPut, "/api/v1/products/"+id, token, repriced); status != 200 {
		t.Fatalf("reprice: %d %s", status, resp.raw)
	}
	status, history, resp := s.priceHistory(t, token, id)
	if status != 200 || len(history) != 1 || history[0]["previous_price"] != "6.00" || history[0]["new_price"] != "4.00" ||
		history[0]["previous_cost"] != "4.20" || history[0]["new_cost"] != "2.50" || history[0]["reason"] != "Promoción" {
		t.Fatalf("history after PUT: %d %s", status, resp.raw)
	}

	repriced["reason"] = strings.Repeat("x", 161)
	status, resp = s.do(t, fiber.MethodPut, "/api/v1/products/"+id, token, repriced)
	expectError(t, status, resp, 400, "VALIDATION_ERROR")
}

func TestPriceHistoryErrors(t *testing.T) {
	s := newTestServer(t)
	id := s.seedProduct(t)
	read := s.token(t, companyA, "products.read")

	status, resp := s.do(t, fiber.MethodGet, "/api/v1/products/"+id+"/price-history", "", nil)
	expectError(t, status, resp, 401, "UNAUTHORIZED")
	status, resp = s.do(t, fiber.MethodGet, "/api/v1/products/"+id+"/price-history", s.token(t, companyA), nil)
	expectError(t, status, resp, 403, "FORBIDDEN")
	status, resp = s.do(t, fiber.MethodGet, "/api/v1/products/"+id+"/price-history", s.token(t, companyB, "products.read"), nil)
	expectError(t, status, resp, 404, "NOT_FOUND")
	status, resp = s.do(t, fiber.MethodGet, "/api/v1/products/"+uuid.NewString()+"/price-history", read, nil)
	expectError(t, status, resp, 404, "NOT_FOUND")
	status, resp = s.do(t, fiber.MethodGet, "/api/v1/products/not-a-uuid/price-history", read, nil)
	expectError(t, status, resp, 400, "BAD_REQUEST")
	if status, history, resp := s.priceHistory(t, read, id); status != 200 || len(history) != 0 {
		t.Fatalf("a product without changes has an empty history: %d %s", status, resp.raw)
	}
}

func TestPatchProductPriceErrors(t *testing.T) {
	s := newTestServer(t)
	id := s.seedProduct(t)
	token := s.token(t, companyA, "products.update")
	path := "/api/v1/products/" + id + "/price"
	tests := []struct {
		name, path, token string
		body              any
		status            int
	}{
		{"below cost", path, token, map[string]any{"sale_price": 2.90}, 400},
		{"negative cost", path, token, map[string]any{"cost": -1}, 400},
		{"zero cost", path, token, map[string]any{"cost": 0}, 400},
		{"precision", path, token, map[string]any{"sale_price": 7.123}, 400},
		{"empty", path, token, map[string]any{}, 400},
		{"null prices", path, token, `{"sale_price":null,"cost":null}`, 400},
		{"invalid decimal", path, token, `{"sale_price":"abc"}`, 400},
		{"malformed JSON", path, token, `{"sale_price":`, 400},
		{"reason length", path, token, map[string]any{"sale_price": 7, "reason": strings.Repeat("x", 161)}, 400},
		{"invalid ID", "/api/v1/products/invalid/price", token, map[string]any{"sale_price": 7}, 400},
		{"not found", "/api/v1/products/" + uuid.NewString() + "/price", token, map[string]any{"sale_price": 7}, 404},
		{"other company", path, s.token(t, companyB, "products.update"), map[string]any{"sale_price": 7}, 404},
		{"no permission", path, s.token(t, companyA), map[string]any{"sale_price": 7}, 403},
		{"no authentication", path, "", map[string]any{"sale_price": 7}, 401},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := s.do(t, fiber.MethodPatch, tt.path, tt.token, tt.body)
			if status != tt.status || resp.Success {
				t.Fatalf("status = %d, want %d; %s", status, tt.status, resp.raw)
			}
		})
	}
	if len(s.store.PriceHistory.Entries[uuid.MustParse(id)]) != 0 {
		t.Fatal("invalid requests created history")
	}
	if !s.store.Products[uuid.MustParse(id)].SalePrice.Equal(decimal.RequireFromString("6.00")) {
		t.Fatal("invalid requests changed the price")
	}
}
