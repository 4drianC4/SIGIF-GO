package response

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

func TestErrorMapsPlainErrorsByStatusCode(t *testing.T) {
	app := fiber.New()
	app.Get("/bad", func(c *fiber.Ctx) error {
		return Error(c, fiber.StatusBadRequest, fmt.Errorf("bad payload"))
	})

	req := httptest.NewRequest(fiber.MethodGet, "/bad", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}

	var body APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error == nil {
		t.Fatal("expected error payload")
	}
	if body.Error.Code != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST code, got %s", body.Error.Code)
	}
}

func TestErrorPreservesWrappedAppError(t *testing.T) {
	app := fiber.New()
	app.Get("/missing", func(c *fiber.Ctx) error {
		return Error(c, fiber.StatusInternalServerError, fmt.Errorf("context: %w", sharedErrors.ErrNotFound))
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/missing", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status %d, got %d", fiber.StatusNotFound, resp.StatusCode)
	}
}
