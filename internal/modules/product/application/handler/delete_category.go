package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
)

func (h *CatalogCommandHandler) HandleDeleteCategory(ctx context.Context, cmd command.DeleteCategory) error {
	return h.service.DeleteCategory(ctx, cmd.CompanyID, cmd.CategoryID)
}
