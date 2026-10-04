package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/service"
)

func (h *CustomerCommandHandler) HandleCreate(ctx context.Context, cmd command.CreateCustomer) (*entity.Customer, error) {
	return h.service.Create(ctx, service.CreateCustomerParams{
		CompanyID:      cmd.CompanyID,
		LegalName:      cmd.LegalName,
		DocumentType:   cmd.DocumentType,
		DocumentNumber: cmd.DocumentNumber,
		Phone:          cmd.Phone,
		Email:          cmd.Email,
	})
}
