package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/user/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

type UserGormRepository struct {
	db *sharedDatabase.Database
}

func NewUserGormRepository(db *sharedDatabase.Database) repository.UserRepository {
	return &UserGormRepository{db: db}
}

func (r *UserGormRepository) Create(ctx context.Context, user *entity.AppUser) error {
	return r.db.GetDB(ctx).Create(mapper.UserToModel(user)).Error
}

func (r *UserGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AppUser, error) {
	var m model.UserModel
	err := r.db.GetDB(ctx).Unscoped().Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	user := mapper.UserToDomain(&m)
	if err := r.attachRoleNames(ctx, []*entity.AppUser{user}); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserGormRepository) GetByEmail(ctx context.Context, email string) (*entity.AppUser, error) {
	var m model.UserModel
	err := r.db.GetDB(ctx).Unscoped().Where("email = ?", email).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	user := mapper.UserToDomain(&m)
	if err := r.attachRoleNames(ctx, []*entity.AppUser{user}); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserGormRepository) List(ctx context.Context, offset, limit int) ([]*entity.AppUser, int64, error) {
	var total int64
	db := r.db.GetDB(ctx).Unscoped().Model(&model.UserModel{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []model.UserModel
	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*entity.AppUser, len(models))
	for i := range models {
		users[i] = mapper.UserToDomain(&models[i])
	}
	if err := r.attachRoleNames(ctx, users); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserGormRepository) Update(ctx context.Context, user *entity.AppUser) error {
	return r.db.GetDB(ctx).Unscoped().Save(mapper.UserToModel(user)).Error
}

func (r *UserGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.GetDB(ctx).Where("id = ?", id).Delete(&model.UserModel{}).Error
}

func (r *UserGormRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.GetDB(ctx).Unscoped().Model(&model.UserModel{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserGormRepository) attachRoleNames(ctx context.Context, users []*entity.AppUser) error {
	if len(users) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(users))
	seen := map[uuid.UUID]struct{}{}
	for _, u := range users {
		if u == nil || u.RoleID == uuid.Nil {
			continue
		}
		if _, ok := seen[u.RoleID]; !ok {
			seen[u.RoleID] = struct{}{}
			ids = append(ids, u.RoleID)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	var roles []model.RoleModel
	if err := r.db.GetDB(ctx).Where("id IN ?", ids).Find(&roles).Error; err != nil {
		return err
	}

	names := make(map[uuid.UUID]string, len(roles))
	for i := range roles {
		names[roles[i].ID] = roles[i].Name
	}
	for _, u := range users {
		if u != nil {
			u.RoleName = names[u.RoleID]
		}
	}
	return nil
}
