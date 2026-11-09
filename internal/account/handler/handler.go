package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/matheuszzy/ledger/internal/account/domain"
	"github.com/matheuszzy/ledger/internal/account/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Router() *gin.Engine {
	router := gin.Default()
	router.HandleMethodNotAllowed = true
	router.POST("/event", h.Event)
	router.GET("/balance", h.Balance)
	router.POST("/reset", h.Reset)
	return router
}

// Event decodes the request and routes it to the matching service method.
func (h *Handler) Event(c *gin.Context) {
	var req eventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, errInvalidBody)
		return
	}

	var result service.EventOutput
	var err error
	switch req.Type {
	case domain.EventTypeDeposit:
		result, err = h.svc.Deposit(req.Destination, req.Amount)
	case domain.EventTypeWithdraw:
		result, err = h.svc.Withdraw(req.Origin, req.Amount)
	case domain.EventTypeTransfer:
		result, err = h.svc.Transfer(req.Origin, req.Destination, req.Amount)
	default:
		err = errUnknownType
	}
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, eventResponse{
		Origin:      toAccountResponse(result.Origin),
		Destination: toAccountResponse(result.Destination),
	})
}

// Balance returns the balance of the account in the account_id query parameter.
func (h *Handler) Balance(c *gin.Context) {
	balance, err := h.svc.Balance(c.Query("account_id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.String(http.StatusOK, strconv.FormatFloat(balance, 'f', -1, 64))
}

// Reset removes every account.
func (h *Handler) Reset(c *gin.Context) {
	h.svc.Reset()
	c.String(http.StatusOK, "OK")
}
