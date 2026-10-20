package dto

import (
	"time"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/repository"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/service"
)

type Movement struct {
	ID         string `json:"id"`
	ProductID  string `json:"product_id"`
	BranchID   int64  `json:"branch_id"`
	LocationID int64  `json:"location_id"`
	Type       string `json:"type"`
	Quantity   string `json:"quantity"`
	Reason     string `json:"reason,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type RegisteredMovement struct {
	Movement Movement `json:"movement"`
	Balance  string   `json:"balance"`
}

type Branch struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Locations []Location `json:"locations"`
}

type Location struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type BalanceItem struct {
	BranchID   int64  `json:"branch_id"`
	LocationID int64  `json:"location_id"`
	Quantity   string `json:"quantity"`
}

func FromMovement(m *entity.Movement) Movement {
	return Movement{
		ID:         m.ID.String(),
		ProductID:  m.ProductID.String(),
		BranchID:   m.BranchID,
		LocationID: m.LocationID,
		Type:       string(m.Type),
		Quantity:   m.Quantity.String(),
		Reason:     m.Reason,
		CreatedAt:  m.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func FromRegisterResult(r *service.RegisterMovementResult) RegisteredMovement {
	return RegisteredMovement{Movement: FromMovement(r.Movement), Balance: r.Balance.String()}
}

func FromMovements(items []*entity.Movement) []Movement {
	result := make([]Movement, len(items))
	for i, m := range items {
		result[i] = FromMovement(m)
	}
	return result
}

func FromBranches(branches []*entity.Branch) []Branch {
	result := make([]Branch, len(branches))
	for i, b := range branches {
		locations := make([]Location, len(b.Locations))
		for j, l := range b.Locations {
			locations[j] = Location{ID: l.ID, Name: l.Name}
		}
		result[i] = Branch{ID: b.ID, Name: b.Name, Locations: locations}
	}
	return result
}

func FromBalances(items []repository.StockBalanceItem) []BalanceItem {
	result := make([]BalanceItem, len(items))
	for i, item := range items {
		result[i] = BalanceItem{BranchID: item.BranchID, LocationID: item.LocationID, Quantity: item.Quantity.String()}
	}
	return result
}
