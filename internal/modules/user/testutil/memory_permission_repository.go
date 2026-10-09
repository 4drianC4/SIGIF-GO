package testutil

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/user/domain/service"
)

var (
	_ repository.UserRepository       = (*MemoryUserRepository)(nil)
	_ repository.RoleRepository       = (*MemoryRoleRepository)(nil)
	_ repository.PermissionRepository = (*MemoryPermissionRepository)(nil)
)

// MemoryUserRepository is a minimal in-memory user repo for the tests that
// only need a UserService (permission catalog endpoints do not touch users).
type MemoryUserRepository struct {
	mu      sync.Mutex
	byID    map[uuid.UUID]*entity.AppUser
	byEmail map[string]*entity.AppUser
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		byID:    map[uuid.UUID]*entity.AppUser{},
		byEmail: map[string]*entity.AppUser{},
	}
}

func (r *MemoryUserRepository) Create(_ context.Context, user *entity.AppUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *MemoryUserRepository) GetByID(_ context.Context, id uuid.UUID) (*entity.AppUser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.byID[id], nil
}

func (r *MemoryUserRepository) GetByEmail(_ context.Context, email string) (*entity.AppUser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.byEmail[email], nil
}

func (r *MemoryUserRepository) List(_ context.Context, _, _ int) ([]*entity.AppUser, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*entity.AppUser, 0, len(r.byID))
	for _, u := range r.byID {
		out = append(out, u)
	}
	return out, int64(len(out)), nil
}

func (r *MemoryUserRepository) Update(_ context.Context, user *entity.AppUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for email, existing := range r.byEmail {
		if existing.ID == user.ID {
			delete(r.byEmail, email)
		}
	}
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *MemoryUserRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byID, id)
	return nil
}

func (r *MemoryUserRepository) ExistsByEmail(_ context.Context, email string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.byEmail[email]
	return ok, nil
}

// MemoryRoleRepository is a minimal in-memory role repo for tests.
type MemoryRoleRepository struct {
	byName map[string]*entity.Role
}

func NewMemoryRoleRepository() *MemoryRoleRepository {
	return &MemoryRoleRepository{byName: map[string]*entity.Role{}}
}

func (r *MemoryRoleRepository) GetByID(_ context.Context, _ uuid.UUID) (*entity.Role, error) {
	return nil, nil
}

func (r *MemoryRoleRepository) GetByName(_ context.Context, name string) (*entity.Role, error) {
	return r.byName[name], nil
}

// MemoryPermissionRepository mirrors the GORM permission repository: ordering
// by module/operation and a unique module.operation code enforced like the
// database index.
type MemoryPermissionRepository struct {
	mu          sync.Mutex
	byCode      map[string]*entity.Permission
	byID        map[uuid.UUID]*entity.Permission
	byRole      map[uuid.UUID]map[uuid.UUID]bool
	Permissions []*entity.Permission
}

func NewMemoryPermissionRepository() *MemoryPermissionRepository {
	return &MemoryPermissionRepository{
		byCode: map[string]*entity.Permission{},
		byID:   map[uuid.UUID]*entity.Permission{},
		byRole: map[uuid.UUID]map[uuid.UUID]bool{},
	}
}

// AssignToRole records that a role holds a permission, so the delete rules can
// be exercised in tests.
func (r *MemoryPermissionRepository) AssignToRole(roleID, permissionID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byRole[roleID] == nil {
		r.byRole[roleID] = map[uuid.UUID]bool{}
	}
	r.byRole[roleID][permissionID] = true
}

// Seed stores a permission directly, as the GORM seed does.
func (r *MemoryPermissionRepository) Seed(p *entity.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byCode[p.Key()]; ok {
		return service.ErrPermissionCodeTaken
	}
	copied := *p
	r.store(&copied)
	return nil
}

func (r *MemoryPermissionRepository) store(p *entity.Permission) {
	r.byCode[p.Key()] = p
	r.byID[p.ID] = p
	r.Permissions = append(r.Permissions, p)
}

func (r *MemoryPermissionRepository) HasPermission(_ context.Context, _ uuid.UUID, module, operation string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.byCode[module+"."+operation]
	return ok, nil
}

func (r *MemoryPermissionRepository) List(_ context.Context, module string, offset, limit int) ([]*entity.Permission, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := make([]*entity.Permission, 0, len(r.byCode))
	for _, p := range r.byCode {
		if module == "" || p.Module == module {
			all = append(all, p)
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Module != all[j].Module {
			return all[i].Module < all[j].Module
		}
		return all[i].Operation < all[j].Operation
	})
	total := int64(len(all))
	from := offset
	if from > len(all) {
		from = len(all)
	}
	to := offset + limit
	if to > len(all) {
		to = len(all)
	}
	return all[from:to], total, nil
}

func (r *MemoryPermissionRepository) GetByModuleOperation(_ context.Context, module, operation string) (*entity.Permission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byCode[module+"."+operation]
	if !ok {
		return nil, nil
	}
	copied := *p
	return &copied, nil
}

func (r *MemoryPermissionRepository) Create(_ context.Context, permission *entity.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byCode[permission.Key()]; ok {
		return service.ErrPermissionCodeTaken
	}
	copied := *permission
	r.store(&copied)
	return nil
}

func (r *MemoryPermissionRepository) GetByID(_ context.Context, id uuid.UUID) (*entity.Permission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	copied := *p
	return &copied, nil
}

func (r *MemoryPermissionRepository) ListAll(_ context.Context, module string) ([]*entity.Permission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := make([]*entity.Permission, 0, len(r.byCode))
	for _, p := range r.byCode {
		if module == "" || p.Module == module {
			all = append(all, p)
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Module != all[j].Module {
			return all[i].Module < all[j].Module
		}
		return all[i].Operation < all[j].Operation
	})
	return all, nil
}

func (r *MemoryPermissionRepository) ListByRole(_ context.Context, roleID uuid.UUID) ([]*entity.Permission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := make([]*entity.Permission, 0, len(r.byRole[roleID]))
	for id := range r.byRole[roleID] {
		if p, ok := r.byID[id]; ok {
			all = append(all, p)
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Module != all[j].Module {
			return all[i].Module < all[j].Module
		}
		return all[i].Operation < all[j].Operation
	})
	return all, nil
}

func (r *MemoryPermissionRepository) Update(_ context.Context, permission *entity.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.byID[permission.ID]
	if !ok {
		return service.ErrPermissionNotFound
	}
	stored.Description = permission.Description
	return nil
}

func (r *MemoryPermissionRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byID[id]
	if !ok {
		return nil
	}
	delete(r.byID, id)
	delete(r.byCode, p.Key())
	for i, item := range r.Permissions {
		if item.ID == id {
			r.Permissions = append(r.Permissions[:i], r.Permissions[i+1:]...)
			break
		}
	}
	return nil
}

func (r *MemoryPermissionRepository) CountRolesByPermission(_ context.Context, permissionID uuid.UUID) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for _, perms := range r.byRole {
		if perms[permissionID] {
			count++
		}
	}
	return count, nil
}
