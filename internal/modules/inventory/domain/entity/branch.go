package entity

import "github.com/google/uuid"

type Branch struct {
	ID        int64
	CompanyID uuid.UUID
	Name      string
	Locations []Location
}

type Location struct {
	ID        int64
	BranchID  int64
	CompanyID uuid.UUID
	Name      string
}
