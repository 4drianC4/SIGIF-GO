package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/jwt"
)

func TestRequireRoles(t *testing.T) {
	jwtManager := jwt.NewManager(&config.Config{JWT: config.JWTConfig{
		Secret:             "super-secret-key",
		AccessTokenExpiry:  60,
		RefreshTokenExpiry: 1440,
		Issuer:             "sigif-test",
	}})

	app := fiber.New()
	app.Use(AuthRequired(jwtManager))
	app.Get("/protected", RequireRoles("manager", "inventory"), func(c *fiber.Ctx) error { return c.SendString("ok") })

	tests := []struct {
		name   string
		roles  []string
		token  bool
		status int
	}{
		{name: "allowed role", roles: []string{"cashier", "inventory"}, token: true, status: fiber.StatusOK},
		{name: "role not allowed", roles: []string{"cashier"}, token: true, status: fiber.StatusForbidden},
		{name: "no roles in token", roles: nil, token: true, status: fiber.StatusForbidden},
		{name: "no token", token: false, status: fiber.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
			if tt.token {
				pair, err := jwtManager.GeneratePair("11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", "user@example.com", tt.roles)
				if err != nil {
					t.Fatalf("GeneratePair returned error: %v", err)
				}
				req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test returned error: %v", err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}
