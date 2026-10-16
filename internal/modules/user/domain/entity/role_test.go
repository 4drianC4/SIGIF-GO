package entity

import "slices"
import "testing"

func TestDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		isSystem bool
		want     string
	}{
		{name: RoleSuperadmin, isSystem: true, want: "Superadministrador"},
		{name: RoleBusinessAdmin, isSystem: true, want: "Administrador de empresa"},
		{name: RoleEmployee, isSystem: true, want: "Empleado"},
		{name: "Vendedor", isSystem: false, want: "Vendedor"},
		{name: "Gerente", isSystem: true, want: "Gerente"},
	}
	for _, tt := range tests {
		if got := DisplayName(tt.name, tt.isSystem); got != tt.want {
			t.Errorf("DisplayName(%q, %t) = %q, want %q", tt.name, tt.isSystem, got, tt.want)
		}
	}
}

func TestRoleDisplayNameMethod(t *testing.T) {
	if got := (&Role{Name: RoleBusinessAdmin, IsSystem: true}).DisplayName(); got != "Administrador de empresa" {
		t.Fatalf("DisplayName() = %q, want the Spanish label", got)
	}
	if got := (&Role{Name: "Vendedor", IsSystem: false}).DisplayName(); got != "Vendedor" {
		t.Fatalf("DisplayName() = %q, want the company name", got)
	}
}

func TestSystemRoleNamesByDisplayPrefix(t *testing.T) {
	tests := []struct {
		q    string
		want []string
	}{
		{q: "empleado", want: []string{RoleEmployee}},
		{q: "EMPRE", want: []string{RoleBusinessAdmin}},
		{q: "administrador de empresa", want: []string{RoleBusinessAdmin}},
		{q: "super", want: []string{RoleSuperadmin}},
		{q: "rol que no existe", want: []string{}},
		{q: "   ", want: []string{}},
		{q: "employ", want: []string{}},
	}
	for _, tt := range tests {
		if got := SystemRoleNamesByDisplayPrefix(tt.q); !slices.Equal(got, tt.want) {
			t.Errorf("SystemRoleNamesByDisplayPrefix(%q) = %v, want %v", tt.q, got, tt.want)
		}
	}
}

func TestAllSystemRoleNames(t *testing.T) {
	names := AllSystemRoleNames()
	want := []string{RoleSuperadmin, RoleBusinessAdmin, RoleEmployee}
	if !slices.Equal(names, want) {
		t.Fatalf("AllSystemRoleNames() = %v, want %v", names, want)
	}
}
