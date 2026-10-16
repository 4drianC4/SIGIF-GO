package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func newOriginGuardApp() *fiber.App {
	app := fiber.New()
	app.Use(OriginGuard([]string{"http://localhost:4200"}))
	app.All("/api/v1/resource", func(c *fiber.Ctx) error { return c.SendString("ok") })
	return app
}

func TestOriginGuard(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		headers map[string]string
		want    int
	}{
		{"post from allowed origin", fiber.MethodPost, map[string]string{"Origin": "http://localhost:4200"}, fiber.StatusOK},
		{"origin match ignores case and trailing slash", fiber.MethodDelete, map[string]string{"Origin": "HTTP://LOCALHOST:4200/"}, fiber.StatusOK},
		{"post from foreign origin", fiber.MethodPost, map[string]string{"Origin": "https://evil.example"}, fiber.StatusForbidden},
		{"patch from opaque origin", fiber.MethodPatch, map[string]string{"Origin": "null"}, fiber.StatusForbidden},
		{"referer fallback allowed", fiber.MethodPut, map[string]string{"Referer": "http://localhost:4200/clientes/1"}, fiber.StatusOK},
		{"referer fallback foreign", fiber.MethodPut, map[string]string{"Referer": "https://evil.example/form"}, fiber.StatusForbidden},
		{"non-browser client without origin", fiber.MethodPost, nil, fiber.StatusOK},
		{"safe method from foreign origin", fiber.MethodGet, map[string]string{"Origin": "https://evil.example"}, fiber.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/v1/resource", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			resp, err := newOriginGuardApp().Test(req)
			if err != nil {
				t.Fatalf("app.Test returned error: %v", err)
			}
			if resp.StatusCode != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, resp.StatusCode)
			}
		})
	}
}
