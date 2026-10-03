package handler

import (
	"context"

	"github.com/sigif/sigif-go/internal/modules/product/application/command"
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/product/domain/event"
	"github.com/sigif/sigif-go/internal/modules/product/domain/service"
)

func (h *CatalogCommandHandler) HandleCreateCategory(ctx context.Context, cmd command.CreateCategory) (*entity.Category, error) {
	category, err := h.service.CreateCategory(ctx, service.CreateCategoryParams{
		TenantID:    cmd.TenantID,
		Name:        cmd.Name,
		Description: cmd.Description,
	})
	if err != nil {
		return nil, err
	}

	_ = h.eventBus.Publish(ctx, event.NewCategoryCreatedEvent(category))
	return category, nil
}
