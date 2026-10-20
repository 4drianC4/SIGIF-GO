package user_test

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/modules/user"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/database"
	"github.com/sigif/sigif-go/internal/shared/validator"
	"go.uber.org/fx"
	"net/http/httptest"
	"testing"
)

func TestUserAndCompanyModuleWiring(t *testing.T) {
	err := fx.ValidateApp(fx.NopLogger, user.Module,
		fx.Supply(fiber.New(), &config.Config{}, &database.Database{}, validator.New()),
		fx.Provide(fx.Annotate(clock.NewRealClock, fx.As(new(clock.Clock)))),
	)
	if err != nil {
		t.Fatal(err)
	}
}

// ValidateApp alone does not execute the route-registration invokes.
func TestCompanyRouteRegisteredByUserModule(t *testing.T) {
	httpApp := fiber.New()
	app := fx.New(fx.NopLogger, user.Module,
		fx.Supply(httpApp, &config.Config{}, &database.Database{}, validator.New()),
		fx.Provide(fx.Annotate(clock.NewRealClock, fx.As(new(clock.Clock)))),
	)
	if err := app.Err(); err != nil {
		t.Fatal(err)
	}
	registered := false
	for _, route := range httpApp.GetRoutes() {
		if route.Method == "GET" && route.Path == "/api/v1/companies" {
			registered = true
		}
	}
	if !registered {
		t.Fatal("company selector route was not registered by Fx")
	}
	response, err := httpApp.Test(httptest.NewRequest("GET", "/api/v1/companies", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 401 {
		t.Fatalf("want 401 from registered protected route, got %d", response.StatusCode)
	}
}
