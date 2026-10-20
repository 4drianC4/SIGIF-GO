package dto

import (
	"time"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type User struct {
	ID         string            `json:"id"`
	CompanyID  *string           `json:"company_id,omitempty"`
	RoleID     string            `json:"role_id"`
	Role       string            `json:"role"`
	FirstName  string            `json:"first_name"`
	LastName   string            `json:"last_name"`
	Username   string            `json:"username"`
	Email      string            `json:"email"`
	Phone      string            `json:"phone,omitempty"`
	Status     entity.UserStatus `json:"status"`
	LastAccess *string           `json:"last_access,omitempty"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}

func FromEntity(u *entity.AppUser) User {
	var companyID *string
	if u.CompanyID != nil {
		s := u.CompanyID.String()
		companyID = &s
	}

	var lastAccess *string
	if u.LastAccess != nil {
		s := u.LastAccess.Format(time.RFC3339)
		lastAccess = &s
	}

	return User{
		ID:         u.ID.String(),
		CompanyID:  companyID,
		RoleID:     u.RoleID.String(),
		Role:       u.RoleName,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Username:   u.Username,
		Email:      u.Email,
		Phone:      u.Phone,
		Status:     u.Status,
		LastAccess: lastAccess,
		CreatedAt:  u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  u.UpdatedAt.Format(time.RFC3339),
	}
}

func FromEntityList(users []*entity.AppUser) []User {
	result := make([]User, len(users))
	for i, u := range users {
		result[i] = FromEntity(u)
	}
	return result
}
