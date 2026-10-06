package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/event"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleCreateCategory(ctx context.Context, cmd command.CreateCategory) (*service.CreatedCategory, error) {
	created, err := h.service.CreateCategory(ctx, service.CreateCategoryParams{
		CompanyID:    cmd.CompanyID,
		ParentID:     cmd.ParentID,
		Name:         cmd.Name,
		Description:  cmd.Description,
		DefaultTax:   cmd.DefaultTax,
		TargetMargin: cmd.TargetMargin,
	})
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewCategoryCreatedEvent(created.Category))
	return created, nil
}
