package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	cashRegisterEntity "github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	appHandler "github.com/sigif/sigif-go/internal/modules/cashsession/application/handler"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
	httpHandler "github.com/sigif/sigif-go/internal/modules/cashsession/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/cashsession/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/modules/cashsession/testutil"
	"github.com/sigif/sigif-go/internal/shared/authcookie"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

const (
	companyA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	companyB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
)

// cashierPerms are the cash session permissions the seed grants to employees.
var cashierPerms = []string{"cash_sessions.open", "cash_sessions.read"}

var openedAt = time.Date(2026, 10, 16, 13, 0, 0, 0, time.UTC)

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details"`
	} `json:"error"`
	raw string
}

type sessionItem struct {
	ID            string          `json:"id"`
	Status        string          `json:"status"`
	OpeningAmount json.RawMessage `json:"opening_amount"`
	OpeningNotes  *string         `json:"opening_notes"`
	OpenedAt      string          `json:"opened_at"`
	ClosedAt      *string         `json:"closed_at"`
	CashierID     string          `json:"cashier_id"`
	AssignmentID  string          `json:"assignment_id"`
	CashRegister  struct {
		ID     string `json:"id"`
		Code   string `json:"code"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"cash_register"`
}

type testServer struct {
	app        *fiber.App
	jwtManager *jwt.JWTManager
	sessions   *testutil.MemorySessionRepository
	registers  *testutil.Registers
	service    *service.CashSessionService
	perms      *stubPermissions
}

type activeSessions struct{}

func (activeSessions) IsActive(context.Context, string) (bool, error) { return true, nil }

type stubPermissions struct {
	mu     sync.Mutex
	byUser map[uuid.UUID]map[string]bool
}

func (p *stubPermissions) HasPermission(_ context.Context, userID uuid.UUID, module, operation string) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.byUser[userID][module+"."+operation], nil
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	jwtManager := jwt.NewManager(&config.Config{JWT: config.JWTConfig{
		Secret:            "test-secret",
		AccessTokenExpiry: 60,
		Issuer:            "sigif-test",
	}})

	sessions := testutil.NewMemorySessionRepository()
	registers := testutil.NewRegisters()
	svc := service.NewCashSessionService(sessions, registers.Access, clock.NewMockClock(openedAt))
	perms := &stubPermissions{byUser: map[uuid.UUID]map[string]bool{}}
	h := httpHandler.NewCashSessionHTTPHandler(
		appHandler.NewCashSessionCommandHandler(svc),
		appHandler.NewCashSessionQueryHandler(svc),
		validator.New(),
	)

	app := fiber.New()
	app.Use(middleware.AuthRequired(jwtManager, activeSessions{}, authcookie.Default()))
	router.RegisterCashSessionRoutes(app.Group("/api/v1"), h, perms)

	return &testServer{app: app, jwtManager: jwtManager, sessions: sessions, registers: registers, service: svc, perms: perms}
}

// user registers a user of the company with the given permissions and returns
// its id and an access token.
func (s *testServer) user(t *testing.T, companyID string, perms ...string) (uuid.UUID, string) {
	t.Helper()
	userID := uuid.New()
	s.perms.mu.Lock()
	s.perms.byUser[userID] = map[string]bool{}
	for _, perm := range perms {
		s.perms.byUser[userID][perm] = true
	}
	s.perms.mu.Unlock()

	token, _, err := s.jwtManager.GenerateAccessToken(userID.String(), uuid.NewString(), companyID, "cajero@sigif.com", "employee")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	return userID, token
}

// cashier returns a cashier of company A holding a new active cash register.
func (s *testServer) cashier(t *testing.T, code string) (userID uuid.UUID, token string, registerID uuid.UUID) {
	t.Helper()
	userID, token = s.user(t, companyA, cashierPerms...)
	register := s.registers.Add(uuid.MustParse(companyA), code, true)
	s.registers.Assign(t, uuid.MustParse(companyA), userID, register.ID)
	return userID, token, register.ID
}

func (s *testServer) send(t *testing.T, method, path, token string, body any) (int, apiResponse) {
	t.Helper()
	var raw []byte
	switch b := body.(type) {
	case nil:
	case string:
		raw = []byte(b)
	default:
		raw, _ = json.Marshal(b)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	var parsed apiResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("response is not JSON (%d): %s", resp.StatusCode, payload)
	}
	parsed.raw = string(payload)
	return resp.StatusCode, parsed
}

func (s *testServer) open(t *testing.T, token string, body any) (int, apiResponse, sessionItem) {
	t.Helper()
	status, resp := s.send(t, fiber.MethodPost, "/api/v1/cash-sessions/", token, body)
	var item sessionItem
	if resp.Success {
		if err := json.Unmarshal(resp.Data, &item); err != nil {
			t.Fatalf("data is not a session: %s", resp.Data)
		}
	}
	return status, resp, item
}

func (s *testServer) current(t *testing.T, token string) (int, apiResponse, sessionItem) {
	t.Helper()
	status, resp := s.send(t, fiber.MethodGet, "/api/v1/cash-sessions/current", token, nil)
	var item sessionItem
	if resp.Success {
		if err := json.Unmarshal(resp.Data, &item); err != nil {
			t.Fatalf("data is not a session: %s", resp.Data)
		}
	}
	return status, resp, item
}

func expectError(t *testing.T, status int, resp apiResponse, wantStatus int, wantCode string) {
	t.Helper()
	if status != wantStatus || resp.Success || resp.Error == nil || resp.Error.Code != wantCode {
		t.Fatalf("got %d %s, want %d %s", status, resp.raw, wantStatus, wantCode)
	}
}

func TestOpenCashSessionFollowsContract(t *testing.T) {
	s := newTestServer(t)
	userID, token, registerID := s.cashier(t, "CAJA-01")

	status, resp, item := s.open(t, token, map[string]any{"opening_amount": 150.5, "opening_notes": " Billetes de 10 "})
	if status != fiber.StatusCreated {
		t.Fatalf("status = %d, body = %s", status, resp.raw)
	}
	if item.Status != "open" || string(item.OpeningAmount) != "150.50" || item.OpeningNotes == nil ||
		*item.OpeningNotes != "Billetes de 10" || item.ClosedAt != nil {
		t.Errorf("session = %s", resp.raw)
	}
	if item.OpenedAt != "2026-10-16T13:00:00Z" || item.CashierID != userID.String() || item.AssignmentID == "" {
		t.Errorf("opened_at / cashier / assignment = %s", resp.raw)
	}
	if item.CashRegister.ID != registerID.String() || item.CashRegister.Code != "CAJA-01" || item.CashRegister.Status != "active" {
		t.Errorf("cash register = %+v", item.CashRegister)
	}
	if !strings.Contains(resp.raw, `"closed_at":null`) {
		t.Errorf("closed_at must be present and null: %s", resp.raw)
	}
}

func TestOpenCashSessionAmounts(t *testing.T) {
	s := newTestServer(t)
	_, token, _ := s.cashier(t, "CAJA-01")
	status, resp, item := s.open(t, token, map[string]any{"opening_amount": 0})
	if status != fiber.StatusCreated || string(item.OpeningAmount) != "0.00" || item.OpeningNotes != nil {
		t.Fatalf("zero opening: %d %s", status, resp.raw)
	}
	if !strings.Contains(resp.raw, `"opening_notes":null`) {
		t.Errorf("opening_notes must be null when not sent: %s", resp.raw)
	}

	tests := []struct {
		name   string
		body   any
		status int
		code   string
		field  string
	}{
		{"negative", map[string]any{"opening_amount": -1}, 400, "VALIDATION_ERROR", "opening_amount"},
		{"three decimals", map[string]any{"opening_amount": "10.005"}, 400, "VALIDATION_ERROR", "opening_amount"},
		{"missing amount", map[string]any{"opening_notes": "x"}, 400, "VALIDATION_ERROR", "opening_amount"},
		{"null amount", `{"opening_amount":null}`, 400, "VALIDATION_ERROR", "opening_amount"},
		{"notes too long", map[string]any{"opening_amount": 1, "opening_notes": strings.Repeat("x", 201)}, 400, "VALIDATION_ERROR", "opening_notes"},
		{"invalid register id", map[string]any{"opening_amount": 1, "cash_register_id": "abc"}, 400, "VALIDATION_ERROR", "cash_register_id"},
		{"amount is text", `{"opening_amount":"abc"}`, 400, "BAD_REQUEST", ""},
		{"malformed JSON", `{"opening_amount":`, 400, "BAD_REQUEST", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(t)
			_, token, _ := s.cashier(t, "CAJA-01")
			status, resp, _ := s.open(t, token, tt.body)
			expectError(t, status, resp, tt.status, tt.code)
			if tt.field != "" {
				if _, ok := resp.Error.Details[tt.field]; !ok {
					t.Errorf("details = %v, want %s", resp.Error.Details, tt.field)
				}
			}
			if s.sessions.OpenCount() != 0 {
				t.Error("an invalid request opened a session")
			}
		})
	}
}

func TestOpenCashSessionConflicts(t *testing.T) {
	s := newTestServer(t)
	_, tokenA, registerID := s.cashier(t, "CAJA-01")
	if status, resp, _ := s.open(t, tokenA, map[string]any{"opening_amount": 100}); status != fiber.StatusCreated {
		t.Fatalf("first opening: %d %s", status, resp.raw)
	}

	status, resp, _ := s.open(t, tokenA, map[string]any{"opening_amount": 100})
	expectError(t, status, resp, 409, "CASH_SESSION_ALREADY_OPEN")

	// A cashier without a cash register: HU-10 answers, no assignment is made.
	_, tokenNoRegister := s.user(t, companyA, cashierPerms...)
	status, resp, _ = s.open(t, tokenNoRegister, map[string]any{"opening_amount": 100})
	expectError(t, status, resp, 404, "NO_ACTIVE_ASSIGNMENT")

	// Sending the cash register of another cashier.
	_, tokenB, _ := s.cashier(t, "CAJA-02")
	status, resp, _ = s.open(t, tokenB, map[string]any{"opening_amount": 100, "cash_register_id": registerID.String()})
	expectError(t, status, resp, 409, "CASH_REGISTER_IN_USE")

	// The cash register went out of service after it was assigned.
	_, tokenC, registerC := s.cashier(t, "CAJA-03")
	s.registers.Repo.Registers[registerC].Status = cashRegisterEntity.CashRegisterStatusInactive
	status, resp, _ = s.open(t, tokenC, map[string]any{"opening_amount": 100})
	expectError(t, status, resp, 409, "CASH_REGISTER_INACTIVE")

	if s.sessions.OpenCount() != 1 {
		t.Errorf("open sessions = %d, want 1", s.sessions.OpenCount())
	}
}

func TestCurrentCashSession(t *testing.T) {
	s := newTestServer(t)
	_, token, _ := s.cashier(t, "CAJA-01")

	status, resp, _ := s.current(t, token)
	expectError(t, status, resp, 404, "NO_OPEN_CASH_SESSION")

	_, _, opened := s.open(t, token, map[string]any{"opening_amount": 80})
	status, resp, item := s.current(t, token)
	if status != fiber.StatusOK || item.ID != opened.ID || item.Status != "open" || item.CashRegister.Code != "CAJA-01" ||
		string(item.OpeningAmount) != "80.00" {
		t.Fatalf("current: %d %s", status, resp.raw)
	}

	// A cashier of another company does not see it.
	_, tokenB := s.user(t, companyB, cashierPerms...)
	status, resp, _ = s.current(t, tokenB)
	expectError(t, status, resp, 404, "NO_OPEN_CASH_SESSION")
}

func TestCashSessionAuthAndPermissions(t *testing.T) {
	s := newTestServer(t)
	routes := []struct {
		method, path, permission string
		body                     any
	}{
		{fiber.MethodPost, "/api/v1/cash-sessions/", "cash_sessions.open", map[string]any{"opening_amount": 10}},
		{fiber.MethodGet, "/api/v1/cash-sessions/current", "cash_sessions.read", nil},
	}
	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			status, resp := s.send(t, r.method, r.path, "", r.body)
			expectError(t, status, resp, 401, "UNAUTHORIZED")

			_, noPerms := s.user(t, companyA)
			status, resp = s.send(t, r.method, r.path, noPerms, r.body)
			expectError(t, status, resp, 403, "FORBIDDEN")

			_, noCompany := s.user(t, "", r.permission)
			status, resp = s.send(t, r.method, r.path, noCompany, r.body)
			expectError(t, status, resp, 400, "BAD_REQUEST")
		})
	}
}
