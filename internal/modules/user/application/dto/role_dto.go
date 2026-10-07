package dto

import (
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

// RoleSummary is the role item served by the role listing (HU-082-02).
type RoleSummary struct {
	ID               string            `json:"id"`
	CompanyID        *string           `json:"company_id"`
	Name             string            `json:"name"`
	Type             entity.RoleType   `json:"type"`
	Status           entity.RoleStatus `json:"status"`
	PermissionsCount int               `json:"permissions_count"`
	Description      string            `json:"description,omitempty"`
	CreatedAt        string            `json:"created_at"`
}

func RoleSummaryFromEntity(r *entity.Role) RoleSummary {
	var companyID *string
	if r.CompanyID != nil {
		s := r.CompanyID.String()
		companyID = &s
	}

	return RoleSummary{
		ID:               r.ID.String(),
		CompanyID:        companyID,
		Name:             r.Name,
		Type:             r.Type(),
		Status:           r.Status,
		PermissionsCount: r.PermissionsCount,
		Description:      r.Description,
		CreatedAt:        r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func RoleSummaryListFromEntity(roles []*entity.Role) []RoleSummary {
	result := make([]RoleSummary, len(roles))
	for i, r := range roles {
		result[i] = RoleSummaryFromEntity(r)
	}
	return result
}
