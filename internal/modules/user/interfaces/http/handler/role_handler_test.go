package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

const (
	companyA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	companyB = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
)

type roleItem struct {
	ID               string  `json:"id"`
	CompanyID        *string `json:"company_id"`
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	Status           string  `json:"status"`
	PermissionsCount int     `json:"permissions_count"`
	Description      string  `json:"description,omitempty"`
	CreatedAt        string  `json:"created_at"`
}

// tokenWithCompany issues a token for the given company; an empty company
// reproduces a user without one (e.g. the global superadmin).
func (s *testServer) tokenWithCompany(t *testing.T, companyID string, perms ...string) string {
	t.Helper()
	userID := uuid.New()
	s.perms.mu.Lock()
	s.perms.byUser[userID] = map[string]bool{}
	for _, perm := range perms {
		s.perms.byUser[userID][perm] = true
	}
	s.perms.mu.Unlock()

	token, _, err := s.jwtManager.GenerateAccessToken(userID.String(), uuid.NewString(), companyID, "admin@sigif.com", "superadmin")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	return token
}

func (s *testServer) listRoles(t *testing.T, query, token string) (int, apiResponse, []roleItem) {
	t.Helper()
	path := "/api/v1/roles/"
	if query != "" {
		path += "?" + query
	}
	status, parsed := s.do(t, fiber.MethodGet, path, "", token)
	var items []roleItem
	if parsed.Success {
		if err := json.Unmarshal(parsed.Data, &items); err != nil {
			t.Fatalf("data is not a role list: %s", parsed.Data)
		}
	}
	return status, parsed, items
}

// seedRole stores a role directly, as the GORM seed does. An empty companyID
// creates a global role.
func (s *testServer) seedRole(t *testing.T, companyID, name string, isSystem bool, status entity.RoleStatus) *entity.Role {
	t.Helper()
	role := &entity.Role{
		ID:          uuid.New(),
		Name:        name,
		Description: "Rol de prueba",
		IsSystem:    isSystem,
		Status:      status,
		CreatedAt:   time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC),
	}
	if companyID != "" {
		id := uuid.MustParse(companyID)
		role.CompanyID = &id
	}
	s.roles.Seed(role)
	return role
}

// grantPermissions assigns the given module.operation permissions to a role,
// registering them first when the seed has not created them yet.
func (s *testServer) grantPermissions(t *testing.T, role *entity.Role, codes ...string) {
	t.Helper()
	for _, code := range codes {
		module, operation, ok := strings.Cut(code, ".")
		if !ok {
			t.Fatalf("permission code %q must be module.operation", code)
		}
		p, err := s.permRepo.GetByModuleOperation(context.Background(), module, operation)
		if err != nil {
			t.Fatalf("GetByModuleOperation(%q): %v", code, err)
		}
		if p == nil {
			p = &entity.Permission{ID: uuid.New(), Module: module, Operation: operation}
			if err := s.permRepo.Seed(p); err != nil {
				t.Fatalf("seed permission %q: %v", code, err)
			}
		}
		s.roles.Grant(role.ID, p.ID)
	}
}

// seedRoleDirectory stores the roles used by most tests: the two system roles,
// two custom roles of company A (one inactive) and one of company B.
func (s *testServer) seedRoleDirectory(t *testing.T) {
	t.Helper()
	superadmin := s.seedRole(t, "", entity.RoleSuperadmin, true, entity.RoleStatusActive)
	s.grantPermissions(t, superadmin, "users.create", "users.list", "customers.list")

	soporte := s.seedRole(t, "", entity.RoleSoporte, true, entity.RoleStatusActive)
	s.grantPermissions(t, soporte, "users.list")

	sellerA := s.seedRole(t, companyA, "Vendedor A", false, entity.RoleStatusActive)
	s.grantPermissions(t, sellerA, "customers.list", "customers.read")

	s.seedRole(t, companyA, "Vendedor Inactivo A", false, entity.RoleStatusInactive)

	sellerB := s.seedRole(t, companyB, "Vendedor B", false, entity.RoleStatusActive)
	s.grantPermissions(t, sellerB, "customers.list")
}

func roleNames(items []roleItem) []string {
	result := make([]string, len(items))
	for i := range items {
		result[i] = items[i].Name
	}
	return result
}

func joinNames(items []roleItem) string {
	return strings.Join(roleNames(items), ",")
}

func TestListRolesEndpointFieldsAndPermissionCount(t *testing.T) {
	s := newTestServer(t)
	s.seedRoleDirectory(t)

	status, resp, items := s.listRoles(t, "", s.tokenWithCompany(t, companyA, "roles.list"))

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if resp.Meta == nil || *resp.Meta != (pageMeta{Page: 1, Limit: 20, Total: 4, TotalPages: 1}) {
		t.Fatalf("meta = %+v, want page 1 limit 20 total 4 total_pages 1", resp.Meta)
	}

	wantCounts := map[string]int{
		"superadmin":          3,
		"soporte":             1,
		"Vendedor A":          2,
		"Vendedor Inactivo A": 0,
	}
	byName := map[string]roleItem{}
	for _, item := range items {
		byName[item.Name] = item
		if item.ID == "" || item.CreatedAt == "" || item.Status == "" || item.Type == "" {
			t.Errorf("incomplete item %+v", item)
		}
	}
	for name, want := range wantCounts {
		item, ok := byName[name]
		if !ok {
			t.Errorf("role %q missing from %v", name, roleNames(items))
			continue
		}
		if item.PermissionsCount != want {
			t.Errorf("%s permissions_count = %d, want %d", name, item.PermissionsCount, want)
		}
		if _, err := uuid.Parse(item.ID); err != nil {
			t.Errorf("%s id %q is not a uuid", name, item.ID)
		}
	}

	t.Run("system roles belong to no company", func(t *testing.T) {
		for _, item := range items {
			if item.Type != "system" {
				continue
			}
			if item.CompanyID != nil {
				t.Errorf("system role %s has company_id %s", item.Name, *item.CompanyID)
			}
		}
	})

	t.Run("custom roles belong to the caller company", func(t *testing.T) {
		for _, item := range items {
			if item.Type != "custom" {
				continue
			}
			if item.CompanyID == nil || *item.CompanyID != companyA {
				t.Errorf("custom role %s has company_id %v, want %s", item.Name, item.CompanyID, companyA)
			}
		}
	})

	t.Run("a role without permissions counts zero", func(t *testing.T) {
		if byName["Vendedor Inactivo A"].PermissionsCount != 0 {
			t.Errorf("permissions_count = %d, want 0", byName["Vendedor Inactivo A"].PermissionsCount)
		}
	})
}

func TestListRolesEndpointSearch(t *testing.T) {
	s := newTestServer(t)
	s.seedRoleDirectory(t)
	token := s.tokenWithCompany(t, companyA, "roles.list")

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{name: "partial name ignoring case", query: "q=VENDE", want: []string{"Vendedor A", "Vendedor Inactivo A"}},
		{name: "partial name lower case", query: "q=inactivo", want: []string{"Vendedor Inactivo A"}},
		{name: "system role", query: "q=super", want: []string{"superadmin"}},
		{name: "surrounding spaces are trimmed", query: "q=%20%20vendedor%20a%20%20", want: []string{"Vendedor A"}},
		{name: "blank search lists everything", query: "q=%20%20%20", want: []string{"soporte", "superadmin", "Vendedor A", "Vendedor Inactivo A"}},
		{name: "search combined with status", query: "q=vendedor&status=inactive", want: []string{"Vendedor Inactivo A"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, items := s.listRoles(t, tt.query, token)
			if status != fiber.StatusOK {
				t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
			}
			if got := joinNames(items); got != strings.Join(tt.want, ",") {
				t.Errorf("names = %v, want %v", roleNames(items), tt.want)
			}
			if resp.Meta.Total != int64(len(tt.want)) {
				t.Errorf("total = %d, want %d", resp.Meta.Total, len(tt.want))
			}
		})
	}
}

func TestListRolesEndpointNoMatchesReturnsEmptyList(t *testing.T) {
	s := newTestServer(t)
	s.seedRoleDirectory(t)

	status, resp, items := s.listRoles(t, "q=no-existe-xyz", s.tokenWithCompany(t, companyA, "roles.list"))

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if string(resp.Data) != "[]" || len(items) != 0 {
		t.Errorf("data = %s, want []", resp.Data)
	}
	if *resp.Meta != (pageMeta{Page: 1, Limit: 20, Total: 0, TotalPages: 0}) {
		t.Errorf("meta = %+v, want total 0", resp.Meta)
	}
}

func TestListRolesEndpointTypeFilter(t *testing.T) {
	s := newTestServer(t)
	s.seedRoleDirectory(t)
	token := s.tokenWithCompany(t, companyA, "roles.list")

	tests := []struct {
		query string
		want  []string
	}{
		{query: "type=system", want: []string{"soporte", "superadmin"}},
		{query: "type=custom", want: []string{"Vendedor A", "Vendedor Inactivo A"}},
		{query: "type=system&q=super", want: []string{"superadmin"}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			status, resp, items := s.listRoles(t, tt.query, token)
			if status != fiber.StatusOK {
				t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
			}
			if got := joinNames(items); got != strings.Join(tt.want, ",") {
				t.Errorf("names = %v, want %v", roleNames(items), tt.want)
			}
			if resp.Meta.Total != int64(len(tt.want)) {
				t.Errorf("total = %d, want %d", resp.Meta.Total, len(tt.want))
			}
		})
	}
}

func TestListRolesEndpointStatusFilter(t *testing.T) {
	s := newTestServer(t)
	s.seedRoleDirectory(t)
	token := s.tokenWithCompany(t, companyA, "roles.list")

	tests := []struct {
		query string
		want  []string
	}{
		{query: "status=active", want: []string{"soporte", "superadmin", "Vendedor A"}},
		{query: "status=inactive", want: []string{"Vendedor Inactivo A"}},
		{query: "status=all", want: []string{"soporte", "superadmin", "Vendedor A", "Vendedor Inactivo A"}},
		{query: "", want: []string{"soporte", "superadmin", "Vendedor A", "Vendedor Inactivo A"}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			status, resp, items := s.listRoles(t, tt.query, token)
			if status != fiber.StatusOK {
				t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
			}
			if got := joinNames(items); got != strings.Join(tt.want, ",") {
				t.Errorf("names = %v, want %v", roleNames(items), tt.want)
			}
			if resp.Meta.Total != int64(len(tt.want)) {
				t.Errorf("total = %d, want %d", resp.Meta.Total, len(tt.want))
			}
		})
	}
}

func TestListRolesEndpointPagination(t *testing.T) {
	s := newTestServer(t)
	for i := 1; i <= 25; i++ {
		s.seedRole(t, companyA, fmt.Sprintf("Rol %02d", i), false, entity.RoleStatusActive)
	}
	token := s.tokenWithCompany(t, companyA, "roles.list")

	tests := []struct {
		query     string
		wantMeta  pageMeta
		wantCount int
		wantFirst string
	}{
		{query: "limit=10", wantMeta: pageMeta{1, 10, 25, 3}, wantCount: 10, wantFirst: "Rol 01"},
		{query: "page=2&limit=10", wantMeta: pageMeta{2, 10, 25, 3}, wantCount: 10, wantFirst: "Rol 11"},
		{query: "page=3&limit=10", wantMeta: pageMeta{3, 10, 25, 3}, wantCount: 5, wantFirst: "Rol 21"},
		{query: "page=4&limit=10", wantMeta: pageMeta{4, 10, 25, 3}, wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			status, resp, items := s.listRoles(t, tt.query, token)
			if status != fiber.StatusOK {
				t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
			}
			if *resp.Meta != tt.wantMeta {
				t.Errorf("meta = %+v, want %+v", *resp.Meta, tt.wantMeta)
			}
			if len(items) != tt.wantCount {
				t.Fatalf("items = %d, want %d", len(items), tt.wantCount)
			}
			if tt.wantFirst != "" && items[0].Name != tt.wantFirst {
				t.Errorf("first = %s, want %s", items[0].Name, tt.wantFirst)
			}
		})
	}
}

func TestListRolesEndpointInvalidParams(t *testing.T) {
	s := newTestServer(t)
	token := s.tokenWithCompany(t, companyA, "roles.list")

	tests := []struct {
		name   string
		query  string
		code   string
		detail string
	}{
		{name: "unknown type", query: "type=xyz", code: "VALIDATION_ERROR", detail: "type"},
		{name: "unknown status", query: "status=xyz", code: "VALIDATION_ERROR", detail: "status"},
		{name: "limit zero", query: "limit=0", code: "VALIDATION_ERROR", detail: "limit"},
		{name: "limit above max", query: "limit=500", code: "VALIDATION_ERROR", detail: "limit"},
		{name: "limit just above max", query: "limit=101", code: "VALIDATION_ERROR", detail: "limit"},
		{name: "page zero", query: "page=0", code: "VALIDATION_ERROR", detail: "page"},
		{name: "negative page", query: "page=-1", code: "VALIDATION_ERROR", detail: "page"},
		{name: "search too long", query: "q=" + strings.Repeat("a", 101), code: "VALIDATION_ERROR", detail: "q"},
		{name: "non numeric page", query: "page=abc", code: "BAD_REQUEST"},
		{name: "non numeric limit", query: "limit=ten", code: "BAD_REQUEST"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.listRoles(t, tt.query, token)
			if status != fiber.StatusBadRequest || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want 400 %s", status, resp.Error, tt.code)
			}
			if tt.detail != "" {
				if _, ok := resp.Error.Details[tt.detail]; !ok {
					t.Errorf("missing detail for %q in %v", tt.detail, resp.Error.Details)
				}
			}
		})
	}
}

func TestListRolesEndpointIsolatesCompanies(t *testing.T) {
	s := newTestServer(t)
	s.seedRoleDirectory(t)

	t.Run("company A never sees company B", func(t *testing.T) {
		_, resp, items := s.listRoles(t, "q=vendedor", s.tokenWithCompany(t, companyA, "roles.list"))
		if resp.Meta.Total != 2 || contains(roleNames(items), "Vendedor B") {
			t.Fatalf("company A got %v", roleNames(items))
		}
		for _, item := range items {
			if item.CompanyID == nil || *item.CompanyID != companyA {
				t.Errorf("unexpected role %+v", item)
			}
		}
	})

	t.Run("company B only sees its own roles plus the system ones", func(t *testing.T) {
		_, resp, items := s.listRoles(t, "", s.tokenWithCompany(t, companyB, "roles.list"))
		if resp.Meta.Total != 3 {
			t.Fatalf("total = %d, want 3 (%v)", resp.Meta.Total, roleNames(items))
		}
		for _, item := range items {
			if item.CompanyID != nil && *item.CompanyID != companyB {
				t.Errorf("leaked role of company %s: %+v", *item.CompanyID, item)
			}
		}
	})

	t.Run("same role name in another company is not matched", func(t *testing.T) {
		s.seedRole(t, companyB, "Vendedor A", false, entity.RoleStatusActive)
		_, resp, items := s.listRoles(t, "q=vendedor%20a", s.tokenWithCompany(t, companyA, "roles.list"))
		if resp.Meta.Total != 1 || items[0].CompanyID == nil || *items[0].CompanyID != companyA {
			t.Fatalf("got %+v, want only company A's role", items)
		}
	})
}

func TestListRolesEndpointWithoutCompanySeesSystemRoles(t *testing.T) {
	s := newTestServer(t)
	s.seedRoleDirectory(t)

	status, resp, items := s.listRoles(t, "", s.tokenWithCompany(t, "", "roles.list"))

	if status != fiber.StatusOK || !resp.Success {
		t.Fatalf("status = %d, want 200 (error %+v)", status, resp.Error)
	}
	if resp.Meta.Total != 2 {
		t.Fatalf("total = %d, want 2 system roles (%v)", resp.Meta.Total, roleNames(items))
	}
	for _, item := range items {
		if item.Type != "system" || item.CompanyID != nil {
			t.Errorf("unexpected role %+v", item)
		}
	}

	t.Run("custom roles of any company stay hidden", func(t *testing.T) {
		if contains(roleNames(items), "Vendedor A") || contains(roleNames(items), "Vendedor B") {
			t.Errorf("custom roles leaked: %v", roleNames(items))
		}
	})
}

func TestListRolesEndpointAuthorization(t *testing.T) {
	s := newTestServer(t)
	s.seedRoleDirectory(t)

	tests := []struct {
		name   string
		token  string
		status int
		code   string
	}{
		{name: "without token", token: "", status: 401, code: "UNAUTHORIZED"},
		{name: "invalid token", token: "not-a-jwt", status: 401, code: "UNAUTHORIZED"},
		{name: "without any permission", token: s.tokenWithCompany(t, companyA), status: 403, code: "FORBIDDEN"},
		{name: "with only users.read", token: s.tokenWithCompany(t, companyA, "users.read"), status: 403, code: "FORBIDDEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp, _ := s.listRoles(t, "", tt.token)
			if status != tt.status || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("got %d %+v, want %d %s", status, resp.Error, tt.status, tt.code)
			}
		})
	}
}

func TestListRolesEndpointDoesNotModifyData(t *testing.T) {
	s := newTestServer(t)
	s.seedRoleDirectory(t)
	before := map[uuid.UUID]entity.Role{}
	for id, r := range s.roles.Roles {
		before[id] = *r
	}
	writes := s.roles.Writes
	token := s.tokenWithCompany(t, companyA, "roles.list")

	for _, query := range []string{"", "q=vendedor", "type=system", "status=inactive&page=2&limit=1", "limit=0"} {
		s.listRoles(t, query, token)
	}

	if s.roles.Writes != writes {
		t.Errorf("listing performed %d writes", s.roles.Writes-writes)
	}
	for id, r := range s.roles.Roles {
		old := before[id]
		if r.Name != old.Name || r.Status != old.Status || r.IsSystem != old.IsSystem || !r.CreatedAt.Equal(old.CreatedAt) {
			t.Errorf("role %s changed: before %+v after %+v", id, old, *r)
		}
	}
}
