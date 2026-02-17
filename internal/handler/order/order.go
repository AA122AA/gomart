package order

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type orderHandler struct {
	lg *zap.Logger
}

func NewOrderHandler(ctx context.Context) *orderHandler {
	return &orderHandler{
		lg: zctx.From(ctx).Named("order handler"),
	}
}

func (oh *orderHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]byte("order get"))
}

func (oh *orderHandler) Create(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]byte("order create"))
}
