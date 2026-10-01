package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/clock"
)

type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Phone        string
	AvatarURL    string
	Roles        []UserRole
	Status       UserStatus
	LastLoginAt  *time.Time
	Settings     UserSettings
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func NewUser(clock clock.Clock, tenantID uuid.UUID, email, passwordHash, firstName, lastName string, roles []UserRole) *User {
	now := clock.NowUTC()
	if len(roles) == 0 {
		roles = []UserRole{RoleViewer}
	}

	return &User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Roles:        roles,
		Status:       UserStatusPending,
		Settings: UserSettings{
			Language:      "es",
			Timezone:      "UTC",
			Theme:         "light",
			Notifications: true,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) HasRole(role UserRole) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (u *User) HasAnyRole(roles []UserRole) bool {
	for _, r := range roles {
		if u.HasRole(r) {
			return true
		}
	}
	return false
}

func (u *User) Update(firstName, lastName, phone, avatarURL string, roles []UserRole, settings UserSettings, clock clock.Clock) {
	u.FirstName = firstName
	u.LastName = lastName
	u.Phone = phone
	u.AvatarURL = avatarURL
	if len(roles) > 0 {
		u.Roles = roles
	}
	u.Settings = settings
	u.UpdatedAt = clock.NowUTC()
}

func (u *User) Activate(clock clock.Clock) {
	u.Status = UserStatusActive
	u.UpdatedAt = clock.NowUTC()
}

func (u *User) Deactivate(clock clock.Clock) {
	u.Status = UserStatusInactive
	u.UpdatedAt = clock.NowUTC()
}

func (u *User) Suspend(clock clock.Clock) {
	u.Status = UserStatusSuspended
	u.UpdatedAt = clock.NowUTC()
}

func (u *User) RecordLogin(clock clock.Clock) {
	now := clock.NowUTC()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

func (u *User) ChangePassword(passwordHash string, clock clock.Clock) {
	u.PasswordHash = passwordHash
	u.UpdatedAt = clock.NowUTC()
}

func (u *User) SoftDelete(clock clock.Clock) {
	now := clock.NowUTC()
	u.DeletedAt = &now
	u.Status = UserStatusInactive
	u.UpdatedAt = now
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}
