package dtos

import (
	"strings"

	"github.com/shopspring/decimal"
)

type RegisterMovementRequest struct {
	ProductID  string           `json:"product_id" validate:"required,uuid"`
	BranchID   int64            `json:"branch_id"`
	LocationID int64            `json:"location_id"`
	Type       string           `json:"type" validate:"required,oneof=in out adjustment"`
	Quantity   *decimal.Decimal `json:"quantity" validate:"required"`
	Reason     string           `json:"reason" validate:"omitempty,max=200"`
}

func (r *RegisterMovementRequest) Normalize() {
	r.ProductID = strings.TrimSpace(r.ProductID)
	r.Type = strings.TrimSpace(r.Type)
	r.Reason = strings.TrimSpace(r.Reason)
}

type ProductStockQuery struct {
	ProductID string `query:"product_id" json:"product_id" validate:"required,uuid"`
}

type ListMovementsQuery struct {
	ProductID string `query:"product_id" json:"product_id" validate:"required,uuid"`
}
