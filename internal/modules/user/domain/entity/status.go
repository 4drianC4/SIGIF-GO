package entity

// UserStatus models the lifecycle of an application user account.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusInvited  UserStatus = "invited"
	UserStatusLocked   UserStatus = "locked"
)

func (s UserStatus) String() string {
	return string(s)
}

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusInvited, UserStatusLocked:
		return true
	default:
		return false
	}
}
