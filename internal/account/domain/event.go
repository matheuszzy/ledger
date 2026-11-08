package domain

type EventType string

const (
	EVENT_TYPE_TRANSFER EventType = "transfer"
	EVENT_TYPE_WITHDRAW EventType = "withdraw"
	EVENT_TYPE_DEPOSIT  EventType = "deposit"
)
