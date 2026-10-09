package testutil

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"

	companyEntity "github.com/sigif/sigif-go/internal/modules/company/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

type MemoryUserRepository struct {
	mu    sync.Mutex
	users map[uuid.UUID]*entity.AppUser
	order []uuid.UUID
}

func NewMemoryUserRepository(users ...*entity.AppUser) *MemoryUserRepository {
	r := &MemoryUserRepository{users: map[uuid.UUID]*entity.AppUser{}}
	for _, u := range users {
		r.users[u.ID] = u
		r.order = append(r.order, u.ID)
	}
	return r
}

func (r *MemoryUserRepository) Create(_ context.Context, user *entity.AppUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	r.order = append(r.order, user.ID)
	return nil
}

func (r *MemoryUserRepository) GetByID(_ context.Context, id uuid.UUID) (*entity.AppUser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.users[id], nil
}

func (r *MemoryUserRepository) GetByEmail(_ context.Context, email string) (*entity.AppUser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (r *MemoryUserRepository) List(_ context.Context, offset, limit int) ([]*entity.AppUser, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := make([]*entity.AppUser, 0, len(r.order))
	for _, id := range r.order {
		if u, ok := r.users[id]; ok {
			all = append(all, u)
		}
	}
	total := int64(len(all))
	if offset > len(all) {
		offset = len(all)
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], total, nil
}

func (r *MemoryUserRepository) Update(_ context.Context, user *entity.AppUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *MemoryUserRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}

func (r *MemoryUserRepository) ExistsByEmail(_ context.Context, email string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

type MemoryHistoryRepository struct {
	mu      sync.Mutex
	records []*entity.UserHistory
	Err     error
}

func NewMemoryHistoryRepository() *MemoryHistoryRepository {
	return &MemoryHistoryRepository{}
}

func (r *MemoryHistoryRepository) Create(_ context.Context, history *entity.UserHistory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return r.Err
	}
	r.records = append(r.records, history)
	return nil
}

func (r *MemoryHistoryRepository) ListByUserID(_ context.Context, userID uuid.UUID, offset, limit int) ([]*entity.UserHistory, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	matches := make([]*entity.UserHistory, 0)
	for i := len(r.records) - 1; i >= 0; i-- {
		if r.records[i].UserID == userID {
			matches = append(matches, r.records[i])
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].CreatedAt.After(matches[j].CreatedAt)
	})
	total := int64(len(matches))
	if offset > len(matches) {
		offset = len(matches)
	}
	end := offset + limit
	if end > len(matches) {
		end = len(matches)
	}
	return matches[offset:end], total, nil
}

func (r *MemoryHistoryRepository) All() []*entity.UserHistory {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*entity.UserHistory(nil), r.records...)
}

type MemoryRoleRepository struct{}

func (MemoryRoleRepository) GetByID(_ context.Context, id uuid.UUID) (*entity.Role, error) {
	return &entity.Role{ID: id}, nil
}

func (MemoryRoleRepository) GetByName(_ context.Context, name string) (*entity.Role, error) {
	return &entity.Role{ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)), Name: name}, nil
}

type MemoryCompanyRepository struct {
	IDs map[uuid.UUID]bool
}

func NewMemoryCompanyRepository(ids ...uuid.UUID) *MemoryCompanyRepository {
	r := &MemoryCompanyRepository{IDs: map[uuid.UUID]bool{}}
	for _, id := range ids {
		r.IDs[id] = true
	}
	return r
}

func (r *MemoryCompanyRepository) ExistsByID(_ context.Context, id uuid.UUID) (bool, error) {
	return r.IDs[id], nil
}

func (r *MemoryCompanyRepository) List(context.Context) ([]companyEntity.Company, error) {
	companies := make([]companyEntity.Company, 0, len(r.IDs))
	for id := range r.IDs {
		companies = append(companies, companyEntity.Company{ID: id})
	}
	return companies, nil
}

type Transactor struct{}

func (Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
