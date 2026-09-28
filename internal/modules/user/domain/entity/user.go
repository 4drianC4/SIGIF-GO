package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/security"
)

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

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusPending   UserStatus = "pending"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID     uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID    *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	Email        string     `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"type:varchar(255);not null"`
	FirstName    string     `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName     string     `json:"last_name" gorm:"type:varchar(100);not null"`
	Phone        string     `json:"phone" gorm:"type:varchar(50)"`
	AvatarURL    string     `json:"avatar_url" gorm:"type:varchar(500)"`
	Roles        []UserRole `json:"roles" gorm:"type:jsonb;not null"`
	Status       UserStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	LastLoginAt  *time.Time `json:"last_login_at" gorm:"index"`
	Settings     UserSettings `json:"settings" gorm:"type:jsonb"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type UserSettings struct {
	Language    string `json:"language" gorm:"default:'es'"`
	Timezone    string `json:"timezone" gorm:"default:'UTC'"`
	Theme       string `json:"theme" gorm:"default:'light'"`
	Notifications bool  `json:"notifications" gorm:"default:true"`
}

func NewUser(clock clock.Clock, tenantID, companyID *uuid.UUID, email, password, firstName, lastName string, roles []UserRole) (*User, error) {
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := clock.NowUTC()
	if len(roles) == 0 {
		roles = []UserRole{RoleViewer}
	}

	return &User{
		TenantID:     *tenantID,
		CompanyID:    companyID,
		Email:        email,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		Roles:        roles,
		Status:       UserStatusPending,
		Settings: UserSettings{
			Language:     "es",
			Timezone:     "UTC",
			Theme:        "light",
			Notifications: true,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
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
	for _, r := range u.Roles {
		for _, check := range roles {
			if r == check {
				return true
			}
		}
	}
	return false
}

func (u *User) CheckPassword(password string) error {
	return security.VerifyPassword(password, u.PasswordHash)
}

func (u *User) ChangePassword(newPassword string) error {
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return nil
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

func (u *User) SoftDelete(clock clock.Clock) {
	now := clock.NowUTC()
	u.DeletedAt = &now
	u.Status = UserStatusInactive
	u.UpdatedAt = now
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

func (User) TableName() string {
	return "users"
}