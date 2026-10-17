package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/inventory/application/command"
	"github.com/sigif/sigif-go/internal/modules/inventory/application/query"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/service"
)

type InventoryCommandHandler struct {
	service *service.InventoryService
}

func NewInventoryCommandHandler(svc *service.InventoryService) *InventoryCommandHandler {
	return &InventoryCommandHandler{service: svc}
}

func (h *InventoryCommandHandler) HandleRegisterMovement(ctx context.Context, cmd command.RegisterMovement) (*service.RegisterMovementResult, error) {
	return h.service.RegisterMovement(ctx, service.RegisterMovementParams{
		CompanyID:  cmd.CompanyID,
		ProductID:  cmd.ProductID,
		BranchID:   cmd.BranchID,
		LocationID: cmd.LocationID,
		Type:       cmd.Type,
		Quantity:   cmd.Quantity,
		Reason:     cmd.Reason,
		UserID:     cmd.UserID,
	})
}

type InventoryQueryHandler struct {
	service *service.InventoryService
}

func NewInventoryQueryHandler(svc *service.InventoryService) *InventoryQueryHandler {
	return &InventoryQueryHandler{service: svc}
}

func (h *InventoryQueryHandler) HandleListBranches(ctx context.Context, q query.ListBranches) ([]*entity.Branch, error) {
	return h.service.ListBranches(ctx, q.CompanyID)
}

func (h *InventoryQueryHandler) HandleProductStock(ctx context.Context, q query.ProductStock) ([]repository.StockBalanceItem, error) {
	return h.service.ProductStock(ctx, q.CompanyID, q.ProductID)
}

func (h *InventoryQueryHandler) HandleListMovements(ctx context.Context, q query.ListMovements) ([]*entity.Movement, int64, error) {
	return h.service.ListMovements(ctx, q.CompanyID, q.ProductID, q.Offset, q.Limit)
}
