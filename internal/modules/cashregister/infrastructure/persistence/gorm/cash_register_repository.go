package gorm

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/cashregister/domain/service"
	"github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/mapper"
	"github.com/sigif/sigif-go/internal/modules/cashregister/infrastructure/persistence/model"
	sharedDatabase "github.com/sigif/sigif-go/internal/shared/database"
)

// activeUserIndex is the unique index that allows one active assignment per
// cashier; any other unique violation on assignments is the per-register one.
const activeUserIndex = "idx_cash_register_assignment_active_user"

type CashRegisterGormRepository struct {
	db *sharedDatabase.Database
}

func NewCashRegisterGormRepository(db *sharedDatabase.Database) repository.CashRegisterRepository {
	return &CashRegisterGormRepository{db: db}
}

func (r *CashRegisterGormRepository) ListRegisters(
	ctx context.Context,
	companyID uuid.UUID,
	filter repository.ListFilter,
	offset, limit int,
) ([]repository.RegisterWithAssignment, int64, error) {
	// Unscoped: the aliased query filters r.deleted_at itself.
	scope := func(db *gorm.DB) *gorm.DB {
		db = db.Unscoped().Table("cash_register AS r").
			Joins("LEFT JOIN cash_register_assignment AS a ON a.cash_register_id = r.cash_register_id AND a.status = ?", string(entity.AssignmentStatusActive)).
			Where("r.company_id = ? AND r.status = ? AND r.deleted_at IS NULL", companyID, string(entity.CashRegisterStatusActive))
		if filter.Availability != nil {
			switch *filter.Availability {
			case entity.AvailabilityAvailable:
				db = db.Where("a.assignment_id IS NULL")
			case entity.AvailabilityInUse:
				db = db.Where("a.assignment_id IS NOT NULL")
			}
		}
		return db
	}

	var total int64
	var registers []model.CashRegisterModel
	err := r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Scopes(scope).Count(&total).Error; err != nil {
			return err
		}
		return tx.Scopes(scope).
			Select("r.*").
			Order("r.code ASC, r.cash_register_id ASC").
			Offset(offset).
			Limit(limit).
			Find(&registers).Error
	})
	if err != nil {
		return nil, 0, err
	}

	ids := make([]uuid.UUID, len(registers))
	for i := range registers {
		ids[i] = registers[i].ID
	}
	var assignments []model.CashRegisterAssignmentModel
	if len(ids) > 0 {
		if err := r.db.GetDB(ctx).
			Where("cash_register_id IN ? AND status = ?", ids, string(entity.AssignmentStatusActive)).
			Find(&assignments).Error; err != nil {
			return nil, 0, err
		}
	}
	byRegister := make(map[uuid.UUID]*entity.Assignment, len(assignments))
	for i := range assignments {
		byRegister[assignments[i].CashRegisterID] = mapper.AssignmentToDomain(&assignments[i])
	}

	items := make([]repository.RegisterWithAssignment, len(registers))
	for i := range registers {
		items[i] = repository.RegisterWithAssignment{
			Register:   mapper.RegisterToDomain(&registers[i]),
			Assignment: byRegister[registers[i].ID],
		}
	}
	return items, total, nil
}

func (r *CashRegisterGormRepository) GetRegister(ctx context.Context, companyID, id uuid.UUID) (*entity.CashRegister, error) {
	var m model.CashRegisterModel
	err := r.db.GetDB(ctx).
		Where("cash_register_id = ? AND company_id = ?", id, companyID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.RegisterToDomain(&m), nil
}

func (r *CashRegisterGormRepository) GetAssignment(ctx context.Context, companyID, id uuid.UUID) (*entity.Assignment, error) {
	return r.findAssignment(ctx, "assignment_id = ? AND company_id = ?", id, companyID)
}

func (r *CashRegisterGormRepository) ActiveAssignmentByRegister(ctx context.Context, companyID, cashRegisterID uuid.UUID) (*entity.Assignment, error) {
	return r.findAssignment(ctx, "cash_register_id = ? AND company_id = ? AND status = ?",
		cashRegisterID, companyID, string(entity.AssignmentStatusActive))
}

func (r *CashRegisterGormRepository) ActiveAssignmentByUser(ctx context.Context, companyID, userID uuid.UUID) (*entity.Assignment, error) {
	return r.findAssignment(ctx, "user_id = ? AND company_id = ? AND status = ?",
		userID, companyID, string(entity.AssignmentStatusActive))
}

func (r *CashRegisterGormRepository) findAssignment(ctx context.Context, query string, args ...any) (*entity.Assignment, error) {
	var m model.CashRegisterAssignmentModel
	err := r.db.GetDB(ctx).Where(query, args...).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.AssignmentToDomain(&m), nil
}

func (r *CashRegisterGormRepository) CreateAssignment(ctx context.Context, assignment *entity.Assignment) error {
	m := mapper.AssignmentToModel(assignment)
	if err := r.db.GetDB(ctx).Omit("CashRegister").Create(m).Error; err != nil {
		if isUniqueViolation(err) {
			if strings.Contains(err.Error(), activeUserIndex) {
				return service.ErrCashierAlreadyAssigned
			}
			return service.ErrCashRegisterInUse
		}
		return err
	}
	return nil
}

// ReleaseAssignment only touches an assignment that is still active, so two
// concurrent releases write once.
func (r *CashRegisterGormRepository) ReleaseAssignment(ctx context.Context, assignment *entity.Assignment) (bool, error) {
	result := r.db.GetDB(ctx).Model(&model.CashRegisterAssignmentModel{}).
		Where("assignment_id = ? AND company_id = ? AND status = ?",
			assignment.ID, assignment.CompanyID, string(entity.AssignmentStatusActive)).
		Updates(map[string]any{
			"status":      string(assignment.Status),
			"released_at": assignment.ReleasedAt,
			"released_by": assignment.ReleasedBy,
			"updated_at":  assignment.UpdatedAt,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *CashRegisterGormRepository) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.GetDB(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(sharedDatabase.WithTx(ctx, tx))
	})
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
