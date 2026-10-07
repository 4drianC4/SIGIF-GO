package gorm

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type RoleGormRepository struct {
	db *sharedDatabase.Database
}

func NewRoleGormRepository(db *sharedDatabase.Database) repository.RoleRepository {
	return &RoleGormRepository{db: db}
}

func (r *RoleGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var m model.RoleModel
	err := r.db.GetDB(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.RoleToDomain(&m), nil
}

func (r *RoleGormRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var m model.RoleModel
	err := r.db.GetDB(ctx).Where("name = ?", name).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.RoleToDomain(&m), nil
}

// List returns a page of the roles the filter makes visible, ordered by name,
// with the permissions of each role counted in the same query (no N+1).
func (r *RoleGormRepository) List(
	ctx context.Context,
	filter repository.RoleListFilter,
	offset, limit int,
) ([]*entity.Role, int64, error) {
	scope := func(db *gorm.DB) *gorm.DB {
		if filter.CompanyID != nil {
			db = db.Where("company_id IS NULL OR company_id = ?", *filter.CompanyID)
		} else {
			db = db.Where("company_id IS NULL")
		}

		if q := strings.TrimSpace(filter.Q); q != "" {
			db = db.Where("LOWER(name) LIKE ?", "%"+escapeLike(strings.ToLower(q))+"%")
		}

		if filter.Type != nil {
			db = db.Where("is_system = ?", *filter.Type == entity.RoleTypeSystem)
		}

		if filter.Status != nil {
			db = db.Where("status = ?", filter.Status.String())
		}
		return db
	}

	var total int64
	if err := r.db.GetDB(ctx).Model(&model.RoleModel{}).Scopes(scope).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []model.RoleListModel
	err := r.db.GetDB(ctx).Model(&model.RoleListModel{}).
		Scopes(scope).
		Select("role.*, (SELECT COUNT(*) FROM role_permission WHERE role_permission.role_id = role.id) AS permissions_count").
		Order("lower(role.name) ASC, role.id ASC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	roles := make([]*entity.Role, len(rows))
	for i := range rows {
		roles[i] = mapper.RoleListToDomain(&rows[i])
	}
	return roles, total, nil
}

// escapeLike makes user input match literally inside a LIKE pattern.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(s)
}
