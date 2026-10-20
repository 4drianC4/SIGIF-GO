package query

import (
	"github.com/google/uuid"
)

type GetCurrentSession struct {
	CompanyID uuid.UUID
	UserID    uuid.UUID
}
