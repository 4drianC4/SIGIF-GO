package dto

import (
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type User struct {
	ID          string              `json:"id"`
	TenantID    string              `json:"tenant_id"`
	Email       string              `json:"email"`
	FirstName   string              `json:"first_name"`
	LastName    string              `json:"last_name"`
	FullName    string              `json:"full_name"`
	Phone       string              `json:"phone,omitempty"`
	AvatarURL   string              `json:"avatar_url,omitempty"`
	Roles       []entity.UserRole   `json:"roles"`
	Status      entity.UserStatus   `json:"status"`
	LastLoginAt *string             `json:"last_login_at,omitempty"`
	Settings    entity.UserSettings `json:"settings"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

func FromEntity(u *entity.User) User {
	var lastLogin *string
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		lastLogin = &s
	}

	return User{
		ID:          u.ID.String(),
		TenantID:    u.TenantID.String(),
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		FullName:    u.FullName(),
		Phone:       u.Phone,
		AvatarURL:   u.AvatarURL,
		Roles:       u.Roles,
		Status:      u.Status,
		LastLoginAt: lastLogin,
		Settings:    u.Settings,
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func FromEntityList(users []*entity.User) []User {
	result := make([]User, len(users))
	for i, u := range users {
		result[i] = FromEntity(u)
	}
	return result
}
