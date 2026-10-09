package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/security"
)

// AppUser is the aggregate root for application user accounts.
type AppUser struct {
	ID                uuid.UUID
	CompanyID         *uuid.UUID
	RoleID            uuid.UUID
	RoleName          string
	FirstName         string
	LastName          string
	Username          string
	Email             string
	Phone             string
	PasswordHash      string
	PasswordAlgorithm string
	RequiresOTP       bool
	Status            UserStatus
	LastAccess        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type RegisterUserParams struct {
	CompanyID *uuid.UUID
	RoleID    uuid.UUID
	FirstName string
	LastName  string
	Email     string
	Password  string
}

func NewUser(clock clock.Clock, params RegisterUserParams) (*AppUser, error) {
	passwordHash, err := security.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	now := clock.NowUTC()
	return &AppUser{
		ID:                uuid.New(),
		CompanyID:         params.CompanyID,
		RoleID:            params.RoleID,
		FirstName:         params.FirstName,
		LastName:          params.LastName,
		Username:          params.Email,
		Email:             params.Email,
		PasswordHash:      passwordHash,
		PasswordAlgorithm: security.Algorithm,
		RequiresOTP:       false,
		Status:            UserStatusActive,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (u *AppUser) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}

// EditUserParams holds the fields an administrator may edit (the same data
// that is provided when registering a user). Nil = leave the field unchanged.
type EditUserParams struct {
	CompanyID *uuid.UUID
	FirstName *string
	LastName  *string
	Email     *string
	RoleID    *uuid.UUID
}

// Edit applies the provided fields onto the user (PATCH semantics).
func (u *AppUser) Edit(params EditUserParams, clock clock.Clock) {
	if params.CompanyID != nil {
		u.CompanyID = params.CompanyID
	}
	if params.FirstName != nil {
		u.FirstName = *params.FirstName
	}
	if params.LastName != nil {
		u.LastName = *params.LastName
	}
	if params.Email != nil {
		u.Email = *params.Email
		u.Username = *params.Email
	}
	if params.RoleID != nil {
		u.RoleID = *params.RoleID
	}
	u.UpdatedAt = clock.NowUTC()
}

func (u *AppUser) Activate(clock clock.Clock) {
	u.Status = UserStatusActive
	u.DeletedAt = nil
	u.UpdatedAt = clock.NowUTC()
}

func (u *AppUser) Deactivate(clock clock.Clock) {
	now := clock.NowUTC()
	u.Status = UserStatusInactive
	u.DeletedAt = &now
	u.UpdatedAt = now
}

func (u *AppUser) IsDeactivated() bool {
	return u.Status == UserStatusInactive
}

func (u *AppUser) Lock(clock clock.Clock) {
	u.Status = UserStatusLocked
	u.UpdatedAt = clock.NowUTC()
}

func (u *AppUser) ChangePassword(password string, clock clock.Clock) error {
	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	u.PasswordHash = passwordHash
	u.PasswordAlgorithm = security.Algorithm
	u.UpdatedAt = clock.NowUTC()
	return nil
}

func (u *AppUser) RecordAccess(clock clock.Clock) {
	now := clock.NowUTC()
	u.LastAccess = &now
	u.UpdatedAt = now
}

func (u *AppUser) SoftDelete(clock clock.Clock) {
	now := clock.NowUTC()
	u.DeletedAt = &now
	u.Status = UserStatusInactive
	u.UpdatedAt = now
}
