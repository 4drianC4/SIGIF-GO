package service

import (
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

const (
	CodeCashRegisterInUse      sharedErrors.ErrorCode = "CASH_REGISTER_IN_USE"
	CodeCashierAlreadyAssigned sharedErrors.ErrorCode = "CASHIER_ALREADY_ASSIGNED"
	CodeNoActiveAssignment     sharedErrors.ErrorCode = "NO_ACTIVE_ASSIGNMENT"
)

var (
	ErrCashRegisterNotFound = sharedErrors.New(sharedErrors.CodeNotFound, "cash register not found", 404)
	ErrAssignmentNotFound   = sharedErrors.New(sharedErrors.CodeNotFound, "cash register assignment not found", 404)
	ErrCashRegisterInUse    = sharedErrors.New(CodeCashRegisterInUse,
		"cash register is already assigned to another cashier", 409)
	ErrCashierAlreadyAssigned = sharedErrors.New(CodeCashierAlreadyAssigned,
		"cashier already has an active cash register; release it first", 409)
	ErrNoActiveAssignment = sharedErrors.New(CodeNoActiveAssignment,
		"cashier has no active cash register assignment", 404)
	ErrReleaseForbidden = sharedErrors.New(sharedErrors.CodeForbidden,
		"only the assigned cashier or a supervisor can release this cash register", 403)
	ErrCashRegisterNotOwned = sharedErrors.New(sharedErrors.CodeForbidden,
		"cash register is not assigned to this cashier", 403)
)
