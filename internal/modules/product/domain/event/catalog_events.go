package event

import (
	"github.com/sigif/sigif-go/internal/modules/product/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/events"
)

type CategoryCreatedEvent struct {
	*events.BaseEvent
	Category *entity.Category
}

func NewCategoryCreatedEvent(category *entity.Category) *CategoryCreatedEvent {
	return &CategoryCreatedEvent{
		BaseEvent: events.NewEvent("category.created", category),
		Category:  category,
	}
}

type ProductCreatedEvent struct {
	*events.BaseEvent
	Product *entity.Product
}

func NewProductCreatedEvent(product *entity.Product) *ProductCreatedEvent {
	return &ProductCreatedEvent{
		BaseEvent: events.NewEvent("product.created", product),
		Product:   product,
	}
}
