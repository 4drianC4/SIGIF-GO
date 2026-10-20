package testutil

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
)

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

type Transactor struct{}

func (Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
