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
	PermCategoriesCreate    = "categories.create"
	PermCategoriesList      = "categories.list"
	PermProductsCreate      = "products.create"
	PermProductsList        = "products.list"
	PermProductsRead        = "products.read"
	PermProductsUpdate      = "products.update"
	PermProductsDelete      = "products.delete"
	PermCustomersCreate     = "customers.create"
	PermCustomersList       = "customers.list"
	PermCustomersRead       = "customers.read"
	PermCustomersUpdate     = "customers.update"
	PermCustomersStatus     = "customers.status"
	PermCustomersDelete     = "customers.delete"
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
		{
			module: "categories",
			ops: []struct{ op, desc string }{
				{"create", "Registrar categorías de productos"},
				{"list", "Listar categorías de productos"},
			},
		},
		{
			module: "products",
			ops: []struct{ op, desc string }{
				{"create", "Registrar productos"},
				{"list", "Listar productos y ver su resumen"},
				{"read", "Ver un producto"},
				{"update", "Editar productos y cambiar estado"},
				{"delete", "Eliminar productos permanentemente"},
			},
		},
		{
			module: "customers",
			ops: []struct{ op, desc string }{
				{"create", "Registrar clientes"},
				{"list", "Listar clientes"},
				{"read", "Ver un cliente"},
				{"update", "Actualizar clientes"},
				{"status", "Cambiar estado de clientes"},
				{"delete", "Eliminar clientes"},
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
