package entity

import "github.com/google/uuid"

type Company struct {
	ID        uuid.UUID
	LegalName string
	TradeName string
}
