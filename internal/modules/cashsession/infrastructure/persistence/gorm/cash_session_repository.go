package gorm

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/cashsession/domain/service"
	"github.com/sigif/sigif-go/internal/modules/cashsession/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/cashsession/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

// openUserIndex is the unique index that allows one open session per cashier;
// any other unique violation on sessions is the per-register one.
const openUserIndex = "idx_cash_session_open_user"

type CashSessionGormRepository struct {
	db *sharedDatabase.Database
}

func NewCashSessionGormRepository(db *sharedDatabase.Database) repository.SessionRepository {
	return &CashSessionGormRepository{db: db}
}

func (r *CashSessionGormRepository) Create(ctx context.Context, session *entity.CashSession) error {
	err := r.db.GetDB(ctx).Omit(clause.Associations).Create(mapper.SessionToModel(session)).Error
	if isUniqueViolation(err) {
		if strings.Contains(err.Error(), openUserIndex) {
			return service.ErrCashierHasOpenSession
		}
		return service.ErrRegisterHasOpenSession
	}
	return err
}

func (r *CashSessionGormRepository) OpenByUser(ctx context.Context, companyID, userID uuid.UUID) (*entity.CashSession, error) {
	return r.findOpen(ctx, "user_id = ? AND company_id = ?", userID, companyID)
}

func (r *CashSessionGormRepository) OpenByRegister(ctx context.Context, companyID, cashRegisterID uuid.UUID) (*entity.CashSession, error) {
	return r.findOpen(ctx, "cash_register_id = ? AND company_id = ?", cashRegisterID, companyID)
}

func (r *CashSessionGormRepository) findOpen(ctx context.Context, query string, args ...any) (*entity.CashSession, error) {
	var m model.CashSessionModel
	err := r.db.GetDB(ctx).
		Where(query, args...).
		Where("status = ?", string(entity.StatusOpen)).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.SessionToDomain(&m), nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
