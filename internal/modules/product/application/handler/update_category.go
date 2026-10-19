package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleUpdateCategory(ctx context.Context, cmd command.UpdateCategory) (*repository.CategoryListItem, error) {
	return h.service.UpdateCategory(ctx, service.UpdateCategoryParams{
		CompanyID:    cmd.CompanyID,
		CategoryID:   cmd.CategoryID,
		Name:         cmd.Name,
		Description:  cmd.Description,
		DefaultTax:   cmd.DefaultTax,
		TargetMargin: cmd.TargetMargin,
	})
}
