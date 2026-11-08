package domain

import "errors"

var ErrNegativeAmount = errors.New("AC001 - Amount cannot be negative")
var ErrInvalidID = errors.New("AC002 - The ID must be a valid integer")
var ErrInsufficientFunds = errors.New("AC003 - insufficient funds")
var ErrSameAccount = errors.New("AC004 - Origin and destination must be different accounts")
