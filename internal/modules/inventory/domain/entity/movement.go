package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MovementType string

const (
	MovementIn         MovementType = "in"
	MovementOut        MovementType = "out"
	MovementAdjustment MovementType = "adjustment"
)

var (
	// ErrMovementTypeInvalid is returned when the movement type is not in the catalog.
	ErrMovementTypeInvalid = errors.New("movement type is not valid")
	// ErrMovementQuantityInvalid is returned when the quantity breaks the rules
	// of its type (non positive for in/out, zero for adjustment).
	ErrMovementQuantityInvalid = errors.New("movement quantity is not valid")
)

// Movement is a signed stock change applied to a product in a branch/location.
// Quantity is stored as the signed delta: positive increases stock, negative
// decreases it.
type Movement struct {
	ID         uuid.UUID
	CompanyID  uuid.UUID
	ProductID  uuid.UUID
	BranchID   int64
	LocationID int64
	Type       MovementType
	Quantity   decimal.Decimal
	Reason     string
	UserID     uuid.UUID
	CreatedAt  time.Time
}

type NewMovementParams struct {
	CompanyID  uuid.UUID
	ProductID  uuid.UUID
	BranchID   int64
	LocationID int64
	Type       MovementType
	Quantity   decimal.Decimal
	Reason     string
	UserID     uuid.UUID
	Now        time.Time
}

// NewMovement validates the type and quantity and normalizes the sign of the
// stored quantity: "in" is positive, "out" is negative and "adjustment" keeps
// the caller's sign (it is the delta itself).
func NewMovement(params NewMovementParams) (*Movement, error) {
	qty := params.Quantity
	switch params.Type {
	case MovementIn:
		if !qty.IsPositive() {
			return nil, ErrMovementQuantityInvalid
		}
	case MovementOut:
		if !qty.IsPositive() {
			return nil, ErrMovementQuantityInvalid
		}
		qty = qty.Neg()
	case MovementAdjustment:
		if qty.IsZero() {
			return nil, ErrMovementQuantityInvalid
		}
	default:
		return nil, ErrMovementTypeInvalid
	}

	return &Movement{
		ID:         uuid.New(),
		CompanyID:  params.CompanyID,
		ProductID:  params.ProductID,
		BranchID:   params.BranchID,
		LocationID: params.LocationID,
		Type:       params.Type,
		Quantity:   qty,
		Reason:     strings.TrimSpace(params.Reason),
		UserID:     params.UserID,
		CreatedAt:  params.Now,
	}, nil
}
