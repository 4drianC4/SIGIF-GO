package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type conflictItem struct {
	Field           string `json:"field"`
	Code            string `json:"code"`
	Value           string `json:"value"`
	ExistingProduct struct {
		ID      string  `json:"id"`
		Name    string  `json:"name"`
		SKU     *string `json:"sku"`
		Barcode *string `json:"barcode"`
		Status  string  `json:"status"`
		Deleted bool    `json:"deleted"`
	} `json:"existing_product"`
}

// conflictsOf decodes details.conflicts of a 409 response.
func conflictsOf(t *testing.T, resp apiResponse) []conflictItem {
	t.Helper()
	raw, _ := json.Marshal(resp.Error.Details["conflicts"])
	var list []conflictItem
	if err := json.Unmarshal(raw, &list); err != nil || len(list) == 0 {
		t.Fatalf("details.conflicts missing in %s", resp.raw)
	}
	return list
}

func TestCreateProductDuplicateContract(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	categoryID := s.createCategory(t, token, map[string]any{"name": "Abarrotes"})
	first := productBody(categoryID)
	first["barcode"] = "7750000000001"
	firstID := s.createProduct(t, token, first)

	body := productBody(categoryID)
	body["name"], body["sku"], body["barcode"] = "Otro producto", "  aba-412 ", "7750000000001"
	status, resp := s.do(t, fiber.MethodPost, "/api/v1/products", token, body)
	expectError(t, status, resp, 409, "CONFLICT")
	if resp.Error.Message != "product has duplicated values" {
		t.Errorf("message = %q", resp.Error.Message)
	}

	list := conflictsOf(t, resp)
	if len(list) != 2 || list[0].Field != "sku" || list[1].Field != "barcode" {
		t.Fatalf("conflicts = %+v, want sku and barcode", list)
	}
	if list[0].Code != "SKU_DUPLICATE" || list[0].Value != "ABA-412" || list[1].Code != "BARCODE_DUPLICATE" || list[1].Value != "7750000000001" {
		t.Errorf("codes and values = %+v", list)
	}
	for _, c := range list {
		p := c.ExistingProduct
		if p.ID != firstID || p.Name != "Galletas María 200g" || p.SKU == nil || *p.SKU != "ABA-412" || p.Status != "active" || p.Deleted {
			t.Errorf("existing product = %+v", p)
		}
	}
	if s.store.ProductCount() != 1 {
		t.Errorf("the conflicting product was stored")
	}
}

func TestUpdateProductDuplicateContract(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	categoryID := s.createCategory(t, token, map[string]any{"name": "Abarrotes"})
	s.createProduct(t, token, productBody(categoryID))
	second := productBody(categoryID)
	second["name"], second["sku"] = "Segundo", "SEG-1"
	secondID := s.createProduct(t, token, second)

	update := updateBody(categoryID)
	update["sku"] = "aba-412"
	status, resp := s.do(t, fiber.MethodPut, "/api/v1/products/"+secondID, token, update)
	expectError(t, status, resp, 409, "CONFLICT")
	if list := conflictsOf(t, resp); len(list) != 1 || list[0].Code != "SKU_DUPLICATE" || list[0].ExistingProduct.Name != "Galletas María 200g" {
		t.Errorf("conflicts = %+v", list)
	}
}

func TestValidateDuplicateListsTheHolder(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	body := productBody(s.createCategory(t, token, map[string]any{"name": "Abarrotes"}))
	body["barcode"] = "7750000000001"
	id := s.createProduct(t, token, body)

	status, resp := s.do(t, fiber.MethodGet, "/api/v1/products/validate-duplicate?sku=aba-412&barcode=7750000000001", token, nil)
	var data struct {
		Exists    bool           `json:"exists"`
		Field     *string        `json:"field"`
		Conflicts []conflictItem `json:"conflicts"`
	}
	if status != 200 || json.Unmarshal(resp.Data, &data) != nil {
		t.Fatalf("got %d %s", status, resp.raw)
	}
	if !data.Exists || data.Field == nil || *data.Field != "sku" || len(data.Conflicts) != 2 || data.Conflicts[1].ExistingProduct.ID != id {
		t.Errorf("data = %s", resp.Data)
	}
}
