package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/matheuszzy/ledger/internal/account/domain"
	"github.com/matheuszzy/ledger/internal/account/repository"
)

var (
	errInvalidBody = errors.New("TE001 - invalid request body")
	errUnknownType = errors.New("TE002 - unknown event type")
)

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		c.String(http.StatusNotFound, "0")
	case errors.Is(err, domain.ErrInsufficientFunds):
		c.String(http.StatusUnprocessableEntity, "0")
	case errors.Is(err, domain.ErrNegativeAmount),
		errors.Is(err, domain.ErrInvalidID),
		errors.Is(err, domain.ErrSameAccount),
		errors.Is(err, errInvalidBody),
		errors.Is(err, errUnknownType):
		c.String(http.StatusBadRequest, err.Error())
	default:
		c.String(http.StatusInternalServerError, "internal error")
	}
}
