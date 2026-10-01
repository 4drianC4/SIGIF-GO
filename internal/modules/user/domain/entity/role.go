package entity

type UserRole string

const (
	RoleSuperAdmin   UserRole = "super_admin"
	RoleTenantAdmin  UserRole = "tenant_admin"
	RoleCompanyAdmin UserRole = "company_admin"
	RoleManager      UserRole = "manager"
	RoleCashier      UserRole = "cashier"
	RoleInventory    UserRole = "inventory"
	RoleSales        UserRole = "sales"
	RoleViewer       UserRole = "viewer"
)

func ParseUserRole(role string) UserRole {
	return UserRole(role)
}