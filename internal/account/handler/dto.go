package handler

import "github.com/matheuszzy/ledger/internal/account/domain"

type accountResponse struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
}

// toAccountResponse converts a domain account to its JSON shape, or nil when absent.
func toAccountResponse(account *domain.Account) *accountResponse {
	if account == nil {
		return nil
	}
	return &accountResponse{ID: account.GetID(), Balance: account.GetBalance().GetAmount()}
}

type eventRequest struct {
	Type        domain.EventType `json:"type"`
	Origin      string           `json:"origin"`
	Destination string           `json:"destination"`
	Amount      float64          `json:"amount"`
}

type eventResponse struct {
	Origin      *accountResponse `json:"origin,omitempty"`
	Destination *accountResponse `json:"destination,omitempty"`
}
