package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleCreateCategory(ctx context.Context, cmd command.CreateCategory) (*entity.Category, error) {
	return h.service.CreateCategory(ctx, service.CreateCategoryParams{
		CompanyID:   cmd.CompanyID,
		Name:        cmd.Name,
		Description: cmd.Description,
	})
}
