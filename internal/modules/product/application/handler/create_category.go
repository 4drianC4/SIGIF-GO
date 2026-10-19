package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleCreateCategory(ctx context.Context, cmd command.CreateCategory) (*service.CreatedCategory, error) {
	return h.service.CreateCategory(ctx, service.CreateCategoryParams{
		CompanyID:    cmd.CompanyID,
		ParentID:     cmd.ParentID,
		Name:         cmd.Name,
		Description:  cmd.Description,
		DefaultTax:   cmd.DefaultTax,
		TargetMargin: cmd.TargetMargin,
	})
}
