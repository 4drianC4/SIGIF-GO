package handler

import (
	"context"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/command"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
)

func (h *SupplierCommandHandler) HandleCreate(ctx context.Context, cmd command.CreateSupplier) (*entity.Supplier, error) {
	return h.service.Create(ctx, cmd.CompanyID, cmd.Fields)
}
func (h *SupplierCommandHandler) HandleUpdate(ctx context.Context, cmd command.UpdateSupplier) (*entity.Supplier, error) {
	return h.service.Update(ctx, cmd.CompanyID, cmd.ID, cmd.Fields)
}
func (h *SupplierCommandHandler) HandleDeactivate(ctx context.Context, cmd command.DeactivateSupplier) (*entity.Supplier, error) {
	return h.service.Deactivate(ctx, cmd.CompanyID, cmd.ID)
}
