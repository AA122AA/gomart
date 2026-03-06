package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AA122AA/gomart.git/internal/db/query"
	"github.com/AA122AA/gomart.git/internal/domain"
	"github.com/AA122AA/gomart.git/internal/service/order"
	"github.com/AA122AA/gomart.git/internal/service/utils"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type OrderService interface {
	Create(ctx context.Context, OID, username string) error
	Get(ctx context.Context, username string) ([]*domain.OrderJSON, error)
	GetAll(ctx context.Context) ([]query.GetAllRow, error)
}

type AuthService interface {
	GetUserByToken(ctx context.Context, rawToken string) (string, error)
}

type orderHandler struct {
	auth AuthService
	srv  OrderService
	lg   *zap.Logger
}

func NewOrderHandler(ctx context.Context, auth AuthService, srv OrderService) *orderHandler {
	return &orderHandler{
		auth: auth,
		srv:  srv,
		lg:   zctx.From(ctx).Named("order handler"),
	}
}

func (oh *orderHandler) GetAll(w http.ResponseWriter, r *http.Request) {

}

func (oh *orderHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := oh.getUser(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		oh.lg.Error("cannot get user", zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	orders, err := oh.srv.Get(r.Context(), user)
	if err != nil {
		oh.lg.Error("cannot get orders", zap.String("user", user), zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func (oh *orderHandler) Create(w http.ResponseWriter, r *http.Request) {

	user, err := oh.getUser(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		oh.lg.Error("cannot get user", zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		oh.lg.Error("cannot read body", zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	err = oh.srv.Create(r.Context(), string(body), user)
	if err != nil {
		switch {
		case errors.Is(err, order.NewErrBadOrderID(nil)):
			http.Error(w, "wrong number", http.StatusBadRequest)
			return
		case errors.Is(err, order.NewErrWrongUser(nil)):
			http.Error(w, "the order number has already been loaded by another user", http.StatusConflict)
			return
		case errors.Is(err, order.NewErrOrderExists(nil)):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]byte("you have already uploaded this order"))
			return
		case errors.Is(err, utils.NewErrWrongLuna(nil)):
			http.Error(w, "you have already uploaded this order", http.StatusUnprocessableEntity)
			return
		default:
			oh.lg.Error("cannot create order", zap.String("order", string(body)), zap.String("user", user), zap.Error(err))
			http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode("order created")
}

func (oh *orderHandler) getUser(ctx context.Context, authHeader string) (string, error) {
	rawToken, err := getToken(authHeader)
	if err != nil {
		return "", fmt.Errorf("cannot get token: %w", err)
	}

	user, err := oh.auth.GetUserByToken(ctx, rawToken)
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
