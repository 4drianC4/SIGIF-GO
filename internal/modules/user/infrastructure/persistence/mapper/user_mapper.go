package mapper

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
)

// ToModel convierte una entidad de dominio a modelo de persistencia.
func ToModel(user *entity.User) *model.UserModel {
	settings, _ := json.Marshal(user.Settings)

	roles := make(datatypes.JSONSlice[string], 0, len(user.Roles))
	for _, r := range user.Roles {
		roles = append(roles, string(r))
	}

	return &model.UserModel{
		ID:           user.ID,
		TenantID:     user.TenantID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Phone:        user.Phone,
		AvatarURL:    user.AvatarURL,
		Roles:        roles,
		Status:       string(user.Status),
		LastLoginAt:  user.LastLoginAt,
		Settings:     settings,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		DeletedAt:    gorm.DeletedAt{Time: timeFromPtr(user.DeletedAt), Valid: user.DeletedAt != nil},
	}
}

// ToDomain convierte un modelo de persistencia a entidad de dominio.
func ToDomain(m *model.UserModel) *entity.User {
	var settings entity.UserSettings
	if len(m.Settings) > 0 {
		_ = json.Unmarshal(m.Settings, &settings)
	}

	roles := make([]entity.UserRole, 0, len(m.Roles))
	for _, r := range m.Roles {
		roles = append(roles, entity.UserRole(r))
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &entity.User{
		ID:           m.ID,
		TenantID:     m.TenantID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		Phone:        m.Phone,
		AvatarURL:    m.AvatarURL,
		Roles:        roles,
		Status:       entity.UserStatus(m.Status),
		LastLoginAt:  m.LastLoginAt,
		Settings:     settings,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}

func timeFromPtr(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}