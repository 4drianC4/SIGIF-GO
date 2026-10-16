package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type InventoryService struct {
	branches  repository.BranchLocationRepository
	balances  repository.StockBalanceRepository
	movements repository.MovementRepository
	products  repository.ProductReader
	clock     clock.Clock
}

func NewInventoryService(
	branches repository.BranchLocationRepository,
	balances repository.StockBalanceRepository,
	movements repository.MovementRepository,
	products repository.ProductReader,
	clk clock.Clock,
) *InventoryService {
	return &InventoryService{
		branches:  branches,
		balances:  balances,
		movements: movements,
		products:  products,
		clock:     clk,
	}
}

type RegisterMovementParams struct {
	CompanyID  uuid.UUID
	ProductID  uuid.UUID
	BranchID   int64
	LocationID int64
	Type       entity.MovementType
	Quantity   decimal.Decimal
	Reason     string
	UserID     uuid.UUID
}

type RegisterMovementResult struct {
	Movement *entity.Movement
	Balance  decimal.Decimal
}

// RegisterMovement validates the request, resolves the branch/location (default
// when omitted), applies the signed delta to the stock_balance cache and records
// the movement, all in one transaction so the balance and the audit log never
// drift.
func (s *InventoryService) RegisterMovement(ctx context.Context, params RegisterMovementParams) (*RegisterMovementResult, error) {
	if params.CompanyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	if params.ProductID == uuid.Nil {
		return nil, ErrProductRequired
	}
	initialStock, exists, err := s.products.Stock(ctx, params.CompanyID, params.ProductID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProductNotFound
	}

	if (params.BranchID <= 0) != (params.LocationID <= 0) {
		return nil, ErrBranchLocationRequired
	}

	movement, err := entity.NewMovement(entity.NewMovementParams{
		CompanyID:  params.CompanyID,
		ProductID:  params.ProductID,
		BranchID:   params.BranchID,
		LocationID: params.LocationID,
		Type:       params.Type,
		Quantity:   params.Quantity,
		Reason:     params.Reason,
		UserID:     params.UserID,
		Now:        s.clock.NowUTC(),
	})
	if err != nil {
		return nil, sharedErrors.Wrap(err, sharedErrors.CodeValidation, "movement is not valid")
	}

	var balance decimal.Decimal
	err = s.balances.WithinTransaction(ctx, func(ctx context.Context) error {
		branchID, locationID := params.BranchID, params.LocationID
		if branchID <= 0 {
			branchID, locationID, err = s.branches.EnsureDefault(ctx, params.CompanyID)
			if err != nil {
				return err
			}
		} else {
			valid, err := s.branches.ExistsLocation(ctx, params.CompanyID, branchID, locationID)
			if err != nil {
				return err
			}
			if !valid {
				return ErrBranchLocationNotFound
			}
		}
		movement.BranchID, movement.LocationID = branchID, locationID
		current, err := s.balances.ApplyDelta(ctx, params.CompanyID, params.ProductID, branchID, locationID, initialStock, movement.Quantity)
		if err != nil {
			return err
		}
		if err := s.movements.Create(ctx, movement); err != nil {
			return err
		}
		if err := s.balances.TouchLastMovement(ctx, params.CompanyID, params.ProductID, movement.CreatedAt); err != nil {
			return err
		}
		balance = current
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RegisterMovementResult{Movement: movement, Balance: balance}, nil
}

func (s *InventoryService) ListBranches(ctx context.Context, companyID uuid.UUID) ([]*entity.Branch, error) {
	if companyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	return s.branches.ListBranches(ctx, companyID)
}

func (s *InventoryService) ProductStock(ctx context.Context, companyID, productID uuid.UUID) ([]repository.StockBalanceItem, error) {
	if companyID == uuid.Nil {
		return nil, ErrCompanyRequired
	}
	if productID == uuid.Nil {
		return nil, ErrProductRequired
	}
	_, found, err := s.products.Stock(ctx, companyID, productID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrProductNotFound
	}
	return s.balances.TotalsByBranch(ctx, companyID, productID)
}

func (s *InventoryService) ListMovements(ctx context.Context, companyID, productID uuid.UUID, offset, limit int) ([]*entity.Movement, int64, error) {
	if companyID == uuid.Nil {
		return nil, 0, ErrCompanyRequired
	}
	if productID == uuid.Nil {
		return nil, 0, ErrProductRequired
	}
	_, found, err := s.products.Stock(ctx, companyID, productID)
	if err != nil {
		return nil, 0, err
	}
	if !found {
		return nil, 0, ErrProductNotFound
	}
	return s.movements.ListByProduct(ctx, companyID, productID, offset, limit)
}
