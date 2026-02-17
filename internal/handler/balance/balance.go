package balance

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type balanceHandler struct {
	lg *zap.Logger
}

func NewBalanceHandler(ctx context.Context) *balanceHandler {
	return &balanceHandler{
		lg: zctx.From(ctx).Named("balance handler"),
	}
}

func (bh *balanceHandler) Get(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]byte("balance get"))
}

func (bh *balanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]byte("balance withdraw"))
}

func (bh *balanceHandler) History(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]byte("balance history"))
}
