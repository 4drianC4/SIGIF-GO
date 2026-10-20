package entity

import (
	"slices"

	"github.com/google/uuid"
)

// Permission is the smallest unit of authorization: a module + an operation.
// Its code is the canonical "module.operation" key (Key).
type Permission struct {
	ID          uuid.UUID
	Module      string
	Operation   string
	Description string
}

// Key returns the unique code that identifies the permission from the backend.
func (p *Permission) Key() string {
	return p.Module + "." + p.Operation
}

// SupportedOperations is the catalog of operations a permission may use. It
// holds the operations defined by HU-082-01 (read, create, update, delete,
// export) plus the ones already seeded for the implemented modules, so a new
// permission never invalidates a permission granted by the seed.
var SupportedOperations = []string{
	"read",
	"list",
	"create",
	"update",
	"delete",
	"export",
	"activate",
	"deactivate",
	"change_password",
	"status",
	"assign",
	"release",
	"release_any",
	"open",
}

// IsSupportedOperation reports whether operation belongs to the catalog.
func IsSupportedOperation(operation string) bool {
	return slices.Contains(SupportedOperations, operation)
}

// ModuleWithOperations pairs a module of the catalog with the operations its
// permissions may use.
type ModuleWithOperations struct {
	Module     string
	Operations []string
}

// SupportedModules returns the catalog of modules a permission may belong to,
// in seed order and without repetitions. The modules come from the same source
// of truth as the seeded permissions.
func SupportedModules() []string {
	var modules []string
	for _, p := range AllPermissions() {
		if !slices.Contains(modules, p.Module) {
			modules = append(modules, p.Module)
		}
	}
	return modules
}

// IsSupportedModule reports whether module belongs to the catalog.
func IsSupportedModule(module string) bool {
	return slices.Contains(SupportedModules(), module)
}

// SupportedModuleCatalog returns every module of the catalog together with the
// operations a permission of that module may use.
func SupportedModuleCatalog() []ModuleWithOperations {
	modules := SupportedModules()
	catalog := make([]ModuleWithOperations, 0, len(modules))
	for _, module := range modules {
		catalog = append(catalog, ModuleWithOperations{
			Module:     module,
			Operations: slices.Clone(SupportedOperations),
		})
	}
	return catalog
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
	PermCategoriesUpdate    = "categories.update"
	PermCategoriesDelete    = "categories.delete"
	PermProductsCreate      = "products.create"
	PermProductsList        = "products.list"
	PermProductsRead        = "products.read"
	PermProductsUpdate      = "products.update"
	PermProductsDelete      = "products.delete"
	PermInventoryCreate     = "inventory.create"
	PermInventoryList       = "inventory.list"
	PermInventoryRead       = "inventory.read"
	PermCustomersCreate     = "customers.create"
	PermCustomersList       = "customers.list"
	PermCustomersRead       = "customers.read"
	PermCustomersUpdate     = "customers.update"
	PermCustomersStatus     = "customers.status"
	PermCustomersDelete     = "customers.delete"
	PermPermissionsRead     = "permissions.read"
	PermPermissionsCreate   = "permissions.create"
	PermRolesList           = "roles.list"
	PermRolesStatus         = "roles.status"
	PermPermissionsUpdate   = "permissions.update"
	PermPermissionsDelete   = "permissions.delete"
	PermPermissionsExport   = "permissions.export"
	PermSuppliersCreate     = "suppliers.create"
	PermSuppliersList       = "suppliers.list"
	PermSuppliersRead       = "suppliers.read"
	PermSuppliersUpdate     = "suppliers.update"
	PermSuppliersDeactivate = "suppliers.deactivate"
)

// Cash register permissions (RF-10).
const (
	PermCashRegistersList       = "cash_registers.list"
	PermCashRegistersAssign     = "cash_registers.assign"
	PermCashRegistersRelease    = "cash_registers.release"
	PermCashRegistersReleaseAny = "cash_registers.release_any"
)

// Cash session permissions (RF-09).
const (
	PermCashSessionsOpen = "cash_sessions.open"
	PermCashSessionsRead = "cash_sessions.read"
)

// IsSystemPermission reports whether the pair module.operation belongs to the
// canonical seed. System permissions are protected: they cannot be deleted.
func IsSystemPermission(module, operation string) bool {
	return slices.ContainsFunc(AllPermissions(), func(p Permission) bool {
		return p.Module == module && p.Operation == operation
	})
}

// SystemPermissionKeys returns the codes (module.operation) of every permission
// of the canonical seed, so clients can tell which ones are protected.
func SystemPermissionKeys() []string {
	perms := AllPermissions()
	keys := make([]string, len(perms))
	for i := range perms {
		keys[i] = perms[i].Key()
	}
	return keys
}

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
				{"update", "Editar categorías de productos"},
				{"delete", "Eliminar categorías de productos"},
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
			module: "inventory",
			ops: []struct{ op, desc string }{
				{"create", "Registrar movimientos de inventario"},
				{"list", "Listar movimientos y sucursales"},
				{"read", "Consultar el stock por sucursal"},
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
		{
			module: "suppliers",
			ops: []struct{ op, desc string }{
				{"create", "Registrar proveedores"},
				{"list", "Listar proveedores y ver su resumen"},
				{"read", "Ver proveedores y sus productos"},
				{"update", "Editar proveedores"},
				{"deactivate", "Desactivar proveedores"},
			},
		},
		{
			module: "permissions",
			ops: []struct{ op, desc string }{
				{"read", "Consultar el catálogo de permisos"},
				{"create", "Registrar permisos"},
				{"update", "Editar la descripción de un permiso"},
				{"delete", "Eliminar permisos"},
				{"export", "Exportar el catálogo de permisos"},
			},
		},
		{
			module: "roles",
			ops: []struct{ op, desc string }{
				{"list", "Listar roles"},
				{"status", "Activar y desactivar roles"},
			},
		},
		{
			module: "cash_registers",
			ops: []struct{ op, desc string }{
				{"list", "Listar cajas y su disponibilidad"},
				{"assign", "Tomar una caja disponible y ver la caja asignada"},
				{"release", "Liberar la caja propia al finalizar el turno"},
				{"release_any", "Liberar la caja de cualquier cajero"},
			},
		},
		{
			module: "cash_sessions",
			ops: []struct{ op, desc string }{
				{"open", "Abrir la sesión de caja con el monto inicial"},
				{"read", "Consultar la sesión de caja abierta propia"},
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
