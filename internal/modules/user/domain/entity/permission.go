package entity

import "github.com/google/uuid"

// Permission is the smallest unit of authorization: a module + an operation.
type Permission struct {
	ID          uuid.UUID
	Module      string
	Operation   string
	Description string
}

func (p *Permission) Key() string {
	return p.Module + "." + p.Operation
}

// Predefined permissions for the modules currently implemented.
const (
	PermUsersCreate         = "users.create"
	PermUsersList           = "users.list"
	PermUsersRead           = "users.read"
	PermUsersUpdate         = "users.update"
	PermUsersDelete         = "users.delete"
	PermUsersActivate       = "users.activate"
	PermUsersDeactivate     = "users.deactivate"
	PermUsersChangePassword = "users.change_password"
)

// AllPermissions returns the canonical set of permissions for the implemented
// modules. Kept as a single source of truth for seeding.
func AllPermissions() []Permission {
	mods := []struct {
		module string
		ops    []struct{ op, desc string }
	}{
		{
			module: "users",
			ops: []struct{ op, desc string }{
				{"create", "Registrar usuarios"},
				{"list", "Listar usuarios"},
				{"read", "Ver un usuario"},
				{"update", "Actualizar usuarios"},
				{"delete", "Eliminar usuarios"},
				{"activate", "Activar usuarios"},
				{"deactivate", "Desactivar usuarios"},
				{"change_password", "Cambiar contraseña de usuarios"},
			},
		},
	}

	var perms []Permission
	for _, m := range mods {
		for _, o := range m.ops {
			perms = append(perms, Permission{
				Module:      m.module,
				Operation:   o.op,
				Description: o.desc,
			})
		}
	}
	return perms
}
