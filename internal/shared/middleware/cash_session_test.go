package middleware

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/errors"
)

type fakeCashSessions struct {
	sessionID uuid.UUID
	err       error
	calls     int
}

func (f *fakeCashSessions) RequireOpenSession(context.Context, uuid.UUID, uuid.UUID) (uuid.UUID, error) {
	f.calls++
	return f.sessionID, f.err
}

var errSessionRequired = errors.New("CASH_SESSION_REQUIRED", "an open cash session is required: open the cash register first", 409)

// guardedApp mounts a fake cash transaction behind RequireOpenCashSession. The
// identity is injected as AuthRequired would do it.
func cashSessionApp(checker CashSessionChecker, userID, companyID uuid.UUID, ran *int, seen *uuid.UUID) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		if userID != uuid.Nil {
			ctx = WithUserID(ctx, userID)
		}
		if companyID != uuid.Nil {
			ctx = WithCompanyID(ctx, companyID)
		}
		c.SetUserContext(ctx)
		return c.Next()
	})
	app.Post("/sales", RequireOpenCashSession(checker), func(c *fiber.Ctx) error {
		*ran++
		*seen, _ = CashSessionIDFromContext(c.UserContext())
		return c.SendStatus(fiber.StatusCreated)
	})
	return app
}

func call(t *testing.T, app *fiber.App) (int, string) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/sales", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(raw, &body)
	return resp.StatusCode, body.Error.Code
}

func TestRequireOpenCashSession(t *testing.T) {
	userID, companyID, sessionID := uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name               string
		user, company      uuid.UUID
		checker            *fakeCashSessions
		wantStatus         int
		wantCode           string
		wantRun, wantCalls int
	}{
		{"with an open session the transaction runs", userID, companyID, &fakeCashSessions{sessionID: sessionID}, 201, "", 1, 1},
		{"without a session it is blocked", userID, companyID, &fakeCashSessions{err: errSessionRequired}, 409, "CASH_SESSION_REQUIRED", 0, 1},
		{"a failing lookup never lets it through", userID, companyID, &fakeCashSessions{err: stderrors.New("db down")}, 500, "INTERNAL_ERROR", 0, 1},
		{"without identity it is 401 and nothing is checked", uuid.Nil, companyID, &fakeCashSessions{}, 401, "UNAUTHORIZED", 0, 0},
		{"without company it is 400 and nothing is checked", userID, uuid.Nil, &fakeCashSessions{}, 400, "BAD_REQUEST", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ran, seen := 0, uuid.Nil
			status, code := call(t, cashSessionApp(tt.checker, tt.user, tt.company, &ran, &seen))
			if status != tt.wantStatus || code != tt.wantCode {
				t.Errorf("got %d %q, want %d %q", status, code, tt.wantStatus, tt.wantCode)
			}
			if ran != tt.wantRun || tt.checker.calls != tt.wantCalls {
				t.Errorf("handler runs = %d, checks = %d; want %d, %d", ran, tt.checker.calls, tt.wantRun, tt.wantCalls)
			}
			if tt.wantRun == 1 && seen != sessionID {
				t.Errorf("session in context = %s, want %s", seen, sessionID)
			}
		})
	}
}
