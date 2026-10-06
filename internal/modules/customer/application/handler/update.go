package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
)

func (h *CustomerCommandHandler) HandleUpdate(ctx context.Context, cmd command.UpdateCustomer) (*entity.Customer, error) {
	return h.service.Update(ctx, service.UpdateCustomerParams{
		ID:             cmd.ID,
		CompanyID:      cmd.CompanyID,
		LegalName:      cmd.LegalName,
		DocumentType:   cmd.DocumentType,
		DocumentNumber: cmd.DocumentNumber,
		Phone:          cmd.Phone,
		Email:          cmd.Email,
		Address:        cmd.Address,
	})
}

func (h *CustomerCommandHandler) HandlePatch(ctx context.Context, cmd command.PatchCustomer) (*entity.Customer, error) {
	return h.service.Patch(ctx, service.PatchCustomerParams{
		ID:             cmd.ID,
		CompanyID:      cmd.CompanyID,
		LegalName:      cmd.LegalName,
		DocumentType:   cmd.DocumentType,
		DocumentNumber: cmd.DocumentNumber,
		Phone:          cmd.Phone,
		Email:          cmd.Email,
		Address:        cmd.Address,
	})
}
