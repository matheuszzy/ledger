package domain

import "errors"

var ErrNegativeAmount = errors.New("AC001 - Amount cannot be negative")
var ErrInvalidID = errors.New("AC002 - The ID must be a valid integer")
