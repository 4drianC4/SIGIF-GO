package entity

type CustomerStatus string

const (
	CustomerStatusActive CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
	CustomerStatusBlocked CustomerStatus = "blocked"
)

func (s CustomerStatus) IsValid() bool {
	switch s {
	case CustomerStatusActive, CustomerStatusInactive, CustomerStatusBlocked:
		return true
	}
	return false
}
