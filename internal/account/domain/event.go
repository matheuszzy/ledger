package domain

type EventType string

const (
	EventTypeTransfer EventType = "transfer"
	EventTypeWithdraw EventType = "withdraw"
	EventTypeDeposit  EventType = "deposit"
)
