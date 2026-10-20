package service

import (
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

const (
	CodeCashSessionAlreadyOpen sharedErrors.ErrorCode = "CASH_SESSION_ALREADY_OPEN"
	CodeNoOpenCashSession      sharedErrors.ErrorCode = "NO_OPEN_CASH_SESSION"
	CodeCashSessionRequired    sharedErrors.ErrorCode = "CASH_SESSION_REQUIRED"
	CodeCashRegisterInactive   sharedErrors.ErrorCode = "CASH_REGISTER_INACTIVE"
)

var (
	ErrCashierHasOpenSession = sharedErrors.New(CodeCashSessionAlreadyOpen,
		"cashier already has an open cash session", 409)
	ErrRegisterHasOpenSession = sharedErrors.New(CodeCashSessionAlreadyOpen,
		"cash register already has an open cash session", 409)
	ErrNoOpenSession = sharedErrors.New(CodeNoOpenCashSession,
		"cashier has no open cash session", 404)
	ErrSessionRequired = sharedErrors.New(CodeCashSessionRequired,
		"an open cash session is required: open the cash register first", 409)
	ErrCashRegisterInactive = sharedErrors.New(CodeCashRegisterInactive,
		"cash register is out of service", 409)
)
