package balance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AA122AA/gomart.git/internal/domain"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type AuthService interface {
	GetUserByToken(ctx context.Context, rawToken string) (string, error)
}

type BalanceService interface {
	Get(ctx context.Context, username string) (*domain.BalanceJSON, error)
	Withdraw(ctx context.Context, username, oid string, amount float32) error
	History(ctx context.Context, username string) ([]*domain.WithdrawJSON, error)
}

type balanceHandler struct {
	auth    AuthService
	balance BalanceService
	lg      *zap.Logger
}

func NewBalanceHandler(ctx context.Context, auth AuthService, balance BalanceService) *balanceHandler {
	return &balanceHandler{
		auth:    auth,
		balance: balance,
		lg:      zctx.From(ctx).Named("balance handler"),
	}
}

func (bh *balanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := bh.getUser(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		bh.lg.Error("cannot get user", zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	res, err := bh.balance.Get(r.Context(), user)
	if err != nil {
		bh.lg.Error("cannot get balance", zap.String("user", user), zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (bh *balanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	user, err := bh.getUser(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		bh.lg.Error("cannot get user", zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()
	type request struct {
		OID string  `json:"order"`
		Sum float32 `json:"sum"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = bh.balance.Withdraw(r.Context(), user, req.OID, req.Sum)
	if err != nil {
		bh.lg.Error("cannot withdraw", zap.String("user", user), zap.String("oid", req.OID), zap.String("sum", fmt.Sprintf("%v", req.Sum)), zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("balance withdraw")
}

func (bh *balanceHandler) History(w http.ResponseWriter, r *http.Request) {
	user, err := bh.getUser(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		bh.lg.Error("cannot get user", zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	res, err := bh.balance.History(r.Context(), user)
	if err != nil {
		bh.lg.Error("cannot get withdraw history", zap.String("user", user), zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (bh *balanceHandler) getUser(ctx context.Context, authHeader string) (string, error) {
	rawToken, err := getToken(authHeader)
	if err != nil {
		return "", fmt.Errorf("cannot get token: %w", err)
	}

	user, err := bh.auth.GetUserByToken(ctx, rawToken)
	if err != nil {
		return "", err
	}

	return user, nil
}

func getToken(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "
	// if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix) {
	if authHeader == "" {
		return "", fmt.Errorf("missing access token")
	}
	rawToken := strings.TrimPrefix(authHeader, bearerPrefix)

	return rawToken, nil
}
