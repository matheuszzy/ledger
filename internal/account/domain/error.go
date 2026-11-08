package domain

import "errors"

var (
	ErrNegativeAmount    = errors.New("AC001 - Amount cannot be negative")
	ErrInvalidID         = errors.New("AC002 - The ID must be a valid integer")
	ErrInsufficientFunds = errors.New("AC003 - insufficient funds")
	ErrSameAccount       = errors.New("AC004 - Origin and destination must be different accounts")
	ErrNotFound          = errors.New("AC005 - Account not found")
)
