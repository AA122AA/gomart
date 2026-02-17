package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/AA122AA/gomart.git/internal/domain"
	authservice "github.com/AA122AA/gomart.git/internal/service/auth"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type UserService interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) (*domain.AccessRefreshJson, error)
	Logout(ctx context.Context, token string) error
	Refresh(ctx context.Context, token string) (*domain.AccessRefreshJson, error)
}

type authHandler struct {
	lg   *zap.Logger
	auth UserService
}

func NewAuthHandler(ctx context.Context, auth UserService) *authHandler {
	return &authHandler{
		lg:   zctx.From(ctx).Named("auth handler"),
		auth: auth,
	}
}

func (ah *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user := domain.LogPassJson{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		ah.lg.Error("cannot decode user json", zap.Error(err))
		return
	}

	err = ah.auth.Register(r.Context(), user.Login, user.Password)
	if err != nil {
		var uerr *authservice.ErrUserExists
		if errors.Is(err, uerr) {
			ah.lg.Error("user already exists", zap.String("user", user.Login), zap.Error(err))
			http.Error(w, "User already exists", http.StatusBadRequest)
			return
		}
		ah.lg.Error("cannot register user", zap.String("user", user.Login), zap.Error(err))
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	res, err := ah.auth.Login(r.Context(), user.Login, user.Password)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		ah.lg.Error("cannot authenticate user", zap.String("user", user.Login), zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(res)
}

func (ah *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user := domain.LogPassJson{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		ah.lg.Error("cannot decode user json", zap.Error(err))
		return
	}

	res, err := ah.auth.Login(r.Context(), user.Login, user.Password)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		ah.lg.Error("cannot authenticate user", zap.String("user", user.Login), zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(res)
}

func (ah *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")

	const bearerPrefix = "Bearer "
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix) {
		http.Error(w, "Missing access token", http.StatusUnauthorized)
		return
	}

	rawToken := strings.TrimPrefix(authHeader, bearerPrefix)
	err := ah.auth.Logout(r.Context(), rawToken)
	if err != nil {
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		ah.lg.Error("cannot log out user", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("auth logout")
}

func (ah *authHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := ah.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		// TODO: если токен невалидный возвращать другой ответ
		ah.lg.Error("cannot refresh tokens", zap.Error(err))
		http.Error(w, "cannot refresh tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
