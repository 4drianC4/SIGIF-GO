package command

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
)

type RegisterMovement struct {
	CompanyID  uuid.UUID
	ProductID  uuid.UUID
	BranchID   int64
	LocationID int64
	Type       entity.MovementType
	Quantity   decimal.Decimal
	Reason     string
	UserID     uuid.UUID
}
