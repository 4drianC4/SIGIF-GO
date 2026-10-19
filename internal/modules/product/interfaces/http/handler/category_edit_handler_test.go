package handler_test

import (
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (s *testServer) newCategory(t *testing.T, token string, body map[string]any) string {
	t.Helper()
	return s.createCategory(t, token, body)
}

func categoryPath(id string) string { return "/api/v1/categories/" + id }

func TestPatchCategoryFollowsContract(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	parent := s.newCategory(t, token, map[string]any{"name": "Abarrotes"})
	id := s.newCategory(t, token, map[string]any{
		"name": "Gaseosas", "parent_id": parent, "description": "Bebidas con gas", "default_tax": "IVA general 13%", "target_margin": 25,
	})
	productBody := productBody(id)
	s.createProduct(t, token, productBody)

	status, resp := s.do(t, fiber.MethodPatch, categoryPath(id), token, map[string]any{"name": "  Bebidas   frías "})
	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, body = %s", status, resp.raw)
	}
	category := decode(t, resp.Data)
	if category["id"] != id || category["name"] != "Bebidas frías" || category["parent_id"] != parent ||
		category["description"] != "Bebidas con gas" || category["default_tax"] != "IVA general 13%" || category["status"] != "active" {
		t.Errorf("category = %s", resp.raw)
	}
	for _, fragment := range []string{`"products_count":1`, `"target_margin":"25"`} {
		if !strings.Contains(resp.raw, fragment) {
			t.Errorf("response missing %s: %s", fragment, resp.raw)
		}
	}

	// null clears what the category can inherit from its parent.
	status, resp = s.do(t, fiber.MethodPatch, categoryPath(id), token, `{"default_tax": null, "target_margin": null, "description": ""}`)
	if status != fiber.StatusOK || !strings.Contains(resp.raw, `"default_tax":null`) || !strings.Contains(resp.raw, `"target_margin":null`) ||
		strings.Contains(resp.raw, `"description"`) || !strings.Contains(resp.raw, `"name":"Bebidas frías"`) {
		t.Fatalf("clearing: %d %s", status, resp.raw)
	}

	status, resp = s.do(t, fiber.MethodPatch, categoryPath(id), token, map[string]any{"default_tax": "Exempt", "target_margin": "12.5", "description": "Nueva"})
	if status != fiber.StatusOK || !strings.Contains(resp.raw, `"default_tax":"Exempt"`) || !strings.Contains(resp.raw, `"target_margin":"12.5"`) ||
		!strings.Contains(resp.raw, `"description":"Nueva"`) {
		t.Fatalf("setting: %d %s", status, resp.raw)
	}
}

func TestPatchCategoryErrors(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	id := s.newCategory(t, token, map[string]any{"name": "Abarrotes"})
	s.newCategory(t, token, map[string]any{"name": "Limpieza"})

	tests := []struct {
		name   string
		path   string
		body   any
		status int
		code   string
		field  string
	}{
		{"empty body", categoryPath(id), "", 400, "VALIDATION_ERROR", ""},
		{"empty object", categoryPath(id), `{}`, 400, "VALIDATION_ERROR", ""},
		{"not an object", categoryPath(id), `[1]`, 400, "BAD_REQUEST", ""},
		{"malformed JSON", categoryPath(id), `{"name":`, 400, "BAD_REQUEST", ""},
		{"unknown field", categoryPath(id), map[string]any{"color": "rojo"}, 400, "VALIDATION_ERROR", "color"},
		{"read-only field", categoryPath(id), map[string]any{"parent_id": uuid.NewString()}, 400, "VALIDATION_ERROR", "parent_id"},
		{"read-only status", categoryPath(id), map[string]any{"status": "inactive"}, 400, "VALIDATION_ERROR", "status"},
		{"name null", categoryPath(id), `{"name": null}`, 400, "VALIDATION_ERROR", "name"},
		{"name blank", categoryPath(id), map[string]any{"name": "   "}, 400, "VALIDATION_ERROR", "name"},
		{"name too short", categoryPath(id), map[string]any{"name": "A"}, 400, "VALIDATION_ERROR", "name"},
		{"name too long", categoryPath(id), map[string]any{"name": strings.Repeat("x", 101)}, 400, "VALIDATION_ERROR", "name"},
		{"name is a number", categoryPath(id), `{"name": 5}`, 400, "VALIDATION_ERROR", "name"},
		{"description too long", categoryPath(id), map[string]any{"description": strings.Repeat("x", 501)}, 400, "VALIDATION_ERROR", "description"},
		{"margin above 100", categoryPath(id), map[string]any{"target_margin": 101}, 400, "VALIDATION_ERROR", "target_margin"},
		{"margin is text", categoryPath(id), `{"target_margin": "abc"}`, 400, "VALIDATION_ERROR", "target_margin"},
		{"several errors at once", categoryPath(id), `{"name": "A", "color": "x"}`, 400, "VALIDATION_ERROR", "color"},
		{"invalid id", categoryPath("abc"), map[string]any{"name": "Otra"}, 400, "BAD_REQUEST", ""},
		{"unknown category", categoryPath(uuid.NewString()), map[string]any{"name": "Otra"}, 404, "NOT_FOUND", ""},
		{"unknown tax", categoryPath(id), map[string]any{"default_tax": "No existe"}, 404, "NOT_FOUND", ""},
		{"repeated name", categoryPath(id), map[string]any{"name": "limpieza"}, 409, "CONFLICT", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := s.do(t, fiber.MethodPatch, tt.path, token, tt.body)
			expectError(t, status, resp, tt.status, tt.code)
			if tt.field != "" {
				if _, ok := resp.Error.Details[tt.field]; !ok {
					t.Errorf("details = %v, want %s", resp.Error.Details, tt.field)
				}
			}
		})
	}
	// Structural problems are reported together; the range of the margin is a
	// domain rule and shows up once they are fixed.
	if status, resp := s.do(t, fiber.MethodPatch, categoryPath(id), token, `{"name": "A", "color": "x"}`); status != 400 || len(resp.Error.Details) != 2 {
		t.Errorf("name and color must be reported together: %s", resp.raw)
	}

	// Another company cannot see the category, nor can an unrelated user edit it.
	status, resp := s.do(t, fiber.MethodPatch, categoryPath(id), s.admin(t, companyB), map[string]any{"name": "Robada"})
	expectError(t, status, resp, 404, "NOT_FOUND")
	status, resp = s.do(t, fiber.MethodGet, "/api/v1/categories", token, nil)
	if status != 200 || !strings.Contains(resp.raw, `"name":"Abarrotes"`) {
		t.Errorf("a failed edit changed the category: %s", resp.raw)
	}
}

func TestDeleteCategoryFollowsContract(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	id := s.newCategory(t, token, map[string]any{"name": "Bebidas"})

	status, resp := s.do(t, fiber.MethodDelete, categoryPath(id), s.admin(t, companyB), nil)
	expectError(t, status, resp, 404, "NOT_FOUND")

	status, resp = s.do(t, fiber.MethodDelete, categoryPath(id), token, nil)
	if status != fiber.StatusNoContent || resp.raw != "" {
		t.Fatalf("delete: got %d %q, want 204 without body", status, resp.raw)
	}
	status, resp = s.do(t, fiber.MethodGet, "/api/v1/categories", token, nil)
	if status != 200 || strings.Contains(resp.raw, id) {
		t.Errorf("the deleted category is still listed: %s", resp.raw)
	}
	status, resp = s.do(t, fiber.MethodDelete, categoryPath(id), token, nil)
	expectError(t, status, resp, 404, "NOT_FOUND")
	status, resp = s.do(t, fiber.MethodPatch, categoryPath(id), token, map[string]any{"name": "Otra"})
	expectError(t, status, resp, 404, "NOT_FOUND")
	status, resp = s.do(t, fiber.MethodDelete, categoryPath("abc"), token, nil)
	expectError(t, status, resp, 400, "BAD_REQUEST")

	// Its name is free again and products can no longer be created in it.
	s.newCategory(t, token, map[string]any{"name": "Bebidas"})
	status, resp = s.do(t, fiber.MethodPost, "/api/v1/products", token, productBody(id))
	expectError(t, status, resp, 404, "NOT_FOUND")
}

func TestDeleteCategoryIsRefusedWithSubcategoriesOrProducts(t *testing.T) {
	s := newTestServer(t)
	token := s.admin(t, companyA)
	parent := s.newCategory(t, token, map[string]any{"name": "Abarrotes"})
	child := s.newCategory(t, token, map[string]any{"name": "Aceites", "parent_id": parent})

	status, resp := s.do(t, fiber.MethodDelete, categoryPath(parent), token, nil)
	expectError(t, status, resp, 409, "CONFLICT")
	if resp.Error.Message != "category has subcategories" {
		t.Errorf("message = %q", resp.Error.Message)
	}

	productID := s.createProduct(t, token, productBody(child))
	status, resp = s.do(t, fiber.MethodDelete, categoryPath(child), token, nil)
	expectError(t, status, resp, 409, "CONFLICT")
	if resp.Error.Message != "category has products" {
		t.Errorf("message = %q", resp.Error.Message)
	}
	// An inactive product still blocks it.
	if status, resp := s.do(t, fiber.MethodPatch, "/api/v1/products/"+productID+"/status", token, map[string]any{"status": "inactive"}); status != 200 {
		t.Fatalf("deactivate: %d %s", status, resp.raw)
	}
	status, resp = s.do(t, fiber.MethodDelete, categoryPath(child), token, nil)
	expectError(t, status, resp, 409, "CONFLICT")

	if status, _ := s.do(t, fiber.MethodDelete, "/api/v1/products/"+productID, token, nil); status != fiber.StatusNoContent {
		t.Fatalf("delete product: %d", status)
	}
	for _, id := range []string{child, parent} {
		if status, resp := s.do(t, fiber.MethodDelete, categoryPath(id), token, nil); status != fiber.StatusNoContent {
			t.Fatalf("delete %s: %d %s", id, status, resp.raw)
		}
	}
}
