package entity

// CustomerStatus representa el estado operativo del cliente.
type CustomerStatus string

const (
	// CustomerStatusActive indica que el cliente está activo y operable.
	CustomerStatusActive CustomerStatus = "active"
	// CustomerStatusInactive indica que el cliente fue desactivado.
	CustomerStatusInactive CustomerStatus = "inactive"
	// CustomerStatusBlocked indica que el cliente fue bloqueado por política.
	CustomerStatusBlocked CustomerStatus = "blocked"
)

// IsValid verifica que el valor sea uno de los estados oficiales.
func (s CustomerStatus) IsValid() bool {
	switch s {
	case CustomerStatusActive, CustomerStatusInactive, CustomerStatusBlocked:
		return true
	}
	return false
}
