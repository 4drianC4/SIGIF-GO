package mapper

import (
	"time"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
)

func UserToModel(u *entity.AppUser) *model.UserModel {
	return &model.UserModel{
		ID:                u.ID,
		CompanyID:         u.CompanyID,
		RoleID:            u.RoleID,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		Username:          u.Username,
		Email:             u.Email,
		Phone:             u.Phone,
		PasswordHash:      u.PasswordHash,
		PasswordAlgorithm: u.PasswordAlgorithm,
		RequiresOTP:       u.RequiresOTP,
		Status:            u.Status.String(),
		LastAccess:        u.LastAccess,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}

func UserToDomain(m *model.UserModel) *entity.AppUser {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &entity.AppUser{
		ID:                m.ID,
		CompanyID:         m.CompanyID,
		RoleID:            m.RoleID,
		FirstName:         m.FirstName,
		LastName:          m.LastName,
		Username:          m.Username,
		Email:             m.Email,
		Phone:             m.Phone,
		PasswordHash:      m.PasswordHash,
		PasswordAlgorithm: m.PasswordAlgorithm,
		RequiresOTP:       m.RequiresOTP,
		Status:            entity.UserStatus(m.Status),
		LastAccess:        m.LastAccess,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		DeletedAt:         deletedAt,
	}
}

func RoleToDomain(m *model.RoleModel) *entity.Role {
	return &entity.Role{
		ID:          m.ID,
		CompanyID:   m.CompanyID,
		Name:        m.Name,
		Description: m.Description,
		IsTemplate:  m.IsTemplate,
		IsSystem:    m.IsSystem,
		Status:      entity.RoleStatus(m.Status),
		CreatedAt:   m.CreatedAt,
	}
}

// RoleListToDomain maps a listing row, including the permissions count
// computed by the query (HU-082-02).
func RoleListToDomain(m *model.RoleListModel) *entity.Role {
	role := RoleToDomain(&m.RoleModel)
	role.PermissionsCount = int(m.PermissionsCount)
	return role
}

func PermissionToModel(p *entity.Permission) *model.PermissionModel {
	return &model.PermissionModel{
		ID:          p.ID,
		Module:      p.Module,
		Operation:   p.Operation,
		Description: p.Description,
	}
}

func PermissionToDomain(m *model.PermissionModel) *entity.Permission {
	return &entity.Permission{
		ID:          m.ID,
		Module:      m.Module,
		Operation:   m.Operation,
		Description: m.Description,
	}
}
