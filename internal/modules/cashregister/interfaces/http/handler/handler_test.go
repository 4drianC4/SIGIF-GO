package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appHandler "github.com/sigif/sigif-go/internal/modules/cashregister/application/handler"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
	httpHandler "github.com/sigif/sigif-go/internal/modules/cashregister/interfaces/http/handler"
	"github.com/sigif/sigif-go/internal/modules/cashregister/interfaces/http/router"
	"github.com/sigif/sigif-go/internal/modules/cashregister/testutil"
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

// cashierPerms are the permissions the seed grants to the employee role.
var cashierPerms = []string{"cash_registers.list", "cash_registers.assign", "cash_registers.release"}

// supervisorPerms are the permissions the seed grants to business_admin.
var supervisorPerms = []string{"cash_registers.list", "cash_registers.release_any"}

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Meta    *struct {
		Page       int   `json:"page"`
		Limit      int   `json:"limit"`
		Total      int64 `json:"total"`
		TotalPages int   `json:"total_pages"`
	} `json:"meta"`
	Error *struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

type cashierItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type registerItem struct {
	ID           string       `json:"id"`
	Code         string       `json:"code"`
	Name         string       `json:"name"`
	Status       string       `json:"status"`
	Availability string       `json:"availability"`
	AssignmentID *string      `json:"assignment_id"`
	AssignedTo   *cashierItem `json:"assigned_to"`
	AssignedAt   *string      `json:"assigned_at"`
}

type assignmentItem struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	CashRegister struct {
		ID     string `json:"id"`
		Code   string `json:"code"`
		Status string `json:"status"`
	} `json:"cash_register"`
	Cashier    cashierItem `json:"cashier"`
	AssignedAt string      `json:"assigned_at"`
	ReleasedAt *string     `json:"released_at"`
	ReleasedBy *string     `json:"released_by"`
}

type testServer struct {
	app        *fiber.App
	jwtManager *jwt.JWTManager
	repo       *testutil.MemoryCashRegisterRepository
	cashiers   *testutil.StubCashierDirectory
	service    *service.CashRegisterService
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
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:            "test-secret",
			AccessTokenExpiry: 60,
			Issuer:            "sigif-test",
		},
		Pagination: config.PaginationConfig{DefaultLimit: 20, MaxLimit: 100},
	}
	jwtManager := jwt.NewManager(cfg)

	repo := testutil.NewMemoryCashRegisterRepository()
	cashiers := testutil.NewStubCashierDirectory()
	clk := clock.NewMockClock(time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC))
	svc := service.NewCashRegisterService(repo, cashiers, clk)
	perms := &stubPermissions{byUser: map[uuid.UUID]map[string]bool{}}
	h := httpHandler.NewCashRegisterHTTPHandler(
		appHandler.NewCashRegisterCommandHandler(svc),
		appHandler.NewCashRegisterQueryHandler(svc),
		perms,
		validator.New(),
		cfg,
	)

	app := fiber.New()
	app.Use(middleware.AuthRequired(jwtManager, activeSessions{}, authcookie.Default()))
	router.RegisterCashRegisterRoutes(app.Group("/api/v1"), h, perms)

	return &testServer{app: app, jwtManager: jwtManager, repo: repo, cashiers: cashiers, service: svc, perms: perms}
}

// user registers a user of the company with the given permissions and returns
// its id and an access token.
func (s *testServer) user(t *testing.T, companyID, name string, perms ...string) (uuid.UUID, string) {
	t.Helper()
	userID := uuid.New()
	s.perms.mu.Lock()
	s.perms.byUser[userID] = map[string]bool{}
	for _, perm := range perms {
		s.perms.byUser[userID][perm] = true
	}
	s.perms.mu.Unlock()
	s.cashiers.Add(userID, name)

	token, _, err := s.jwtManager.GenerateAccessToken(userID.String(), uuid.NewString(), companyID, "cajero@sigif.com", "employee")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	return userID, token
}

func (s *testServer) cashier(t *testing.T, companyID, name string) (uuid.UUID, string) {
	t.Helper()
	return s.user(t, companyID, name, cashierPerms...)
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
	return resp.StatusCode, parsed
}

func (s *testServer) list(t *testing.T, query, token string) (int, apiResponse, []registerItem) {
	t.Helper()
	path := "/api/v1/cash-registers/"
	if query != "" {
		path += "?" + query
	}
	status, resp := s.send(t, fiber.MethodGet, path, token, nil)
	var items []registerItem
	if resp.Success {
		if err := json.Unmarshal(resp.Data, &items); err != nil {
			t.Fatalf("data is not a list: %s", resp.Data)
		}
	}
	return status, resp, items
}

func (s *testServer) assign(t *testing.T, token string, body any) (int, apiResponse, assignmentItem) {
	t.Helper()
	status, resp := s.send(t, fiber.MethodPost, "/api/v1/cash-register-assignments/", token, body)
	return status, resp, decodeAssignment(resp)
}

func (s *testServer) assignRegister(t *testing.T, token string, registerID uuid.UUID) (int, apiResponse, assignmentItem) {
	t.Helper()
	return s.assign(t, token, map[string]any{"cash_register_id": registerID.String()})
}

func (s *testServer) current(t *testing.T, token string) (int, apiResponse, assignmentItem) {
	t.Helper()
	status, resp := s.send(t, fiber.MethodGet, "/api/v1/cash-register-assignments/current", token, nil)
	return status, resp, decodeAssignment(resp)
}

func (s *testServer) release(t *testing.T, id, token string) (int, apiResponse, assignmentItem) {
	t.Helper()
	status, resp := s.send(t, fiber.MethodPost, "/api/v1/cash-register-assignments/"+id+"/release", token, nil)
	return status, resp, decodeAssignment(resp)
}

func decodeAssignment(resp apiResponse) assignmentItem {
	var a assignmentItem
	if resp.Success {
		_ = json.Unmarshal(resp.Data, &a)
	}
	return a
}

func byCode(items []registerItem) map[string]registerItem {
	result := make(map[string]registerItem, len(items))
	for _, item := range items {
		result[item.Code] = item
	}
	return result
}

func codes(items []registerItem) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.Code
	}
	return result
}

func errorCode(resp apiResponse) string {
	if resp.Error == nil {
		return ""
	}
	return resp.Error.Code
}
