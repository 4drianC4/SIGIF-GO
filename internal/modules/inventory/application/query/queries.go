package query

import "github.com/google/uuid"

type ListBranches struct {
	CompanyID uuid.UUID
}

type ProductStock struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
}

type ListMovements struct {
	CompanyID uuid.UUID
	ProductID uuid.UUID
	Offset    int
	Limit     int
}
