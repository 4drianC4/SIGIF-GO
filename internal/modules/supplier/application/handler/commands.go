package handler

import (
	"context"
	"github.com/sigif/sigif-go/internal/modules/supplier/application/command"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/supplier/domain/service"
)

func (h *SupplierCommandHandler) HandleCreate(ctx context.Context, cmd command.CreateSupplier) (*entity.Supplier, error) {
	return h.service.Create(ctx, cmd.CompanyID, cmd.Fields)
}

func (h *SupplierCommandHandler) HandleAssociate(ctx context.Context, cmd command.AssociateProduct) (*entity.ProductSupplier, error) {
	return h.service.Associate(ctx, service.AssociationParams(cmd))
}
func (h *SupplierCommandHandler) HandleAssociateProducts(ctx context.Context, cmd command.AssociateProducts) ([]*entity.ProductSupplier, error) {
	return h.service.AssociateProducts(ctx, cmd.CompanyID, cmd.SupplierID, cmd.ProductIDs)
}
func (h *SupplierCommandHandler) HandleUpdateAssociation(ctx context.Context, cmd command.UpdateProductAssociation) (*entity.ProductSupplier, error) {
	return h.service.UpdateAssociation(ctx, service.AssociationParams(cmd))
}
func (h *SupplierCommandHandler) HandleUnlink(ctx context.Context, cmd command.UnlinkProduct) error {
	return h.service.Unlink(ctx, cmd.CompanyID, cmd.SupplierID, cmd.ProductID)
}
func (h *SupplierCommandHandler) HandleUpdate(ctx context.Context, cmd command.UpdateSupplier) (*entity.Supplier, error) {
	return h.service.Update(ctx, cmd.CompanyID, cmd.ID, cmd.Fields)
}
func (h *SupplierCommandHandler) HandleDeactivate(ctx context.Context, cmd command.DeactivateSupplier) (*entity.Supplier, error) {
	return h.service.Deactivate(ctx, cmd.CompanyID, cmd.ID)
}
