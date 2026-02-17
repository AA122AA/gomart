package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/AA122AA/gomart.git/internal/config"
	"github.com/AA122AA/gomart.git/internal/middleware"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type AuthService interface {
	VerifyAccessToken(ctx context.Context, rawToken string) error
}

type authHandler interface {
	Refresh(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type orderHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
}

type balanceHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	Withdraw(w http.ResponseWriter, r *http.Request)
	History(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	srv *http.Server
	lg  *zap.Logger
}

func NewServer(ctx context.Context, router http.Handler, cfg *config.Config) *Server {
	return &Server{
		srv: &http.Server{
			Addr:    cfg.HostAddr,
			Handler: router,
		},
		lg: zctx.From(ctx).Named("server"),
	}
}

func (s *Server) OnShutDown(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	<-ctx.Done()
	if s.srv != nil {
		if err := s.srv.Shutdown(ctx); err != nil {
			s.lg.Fatal("failed to shutdown server", zap.Error(err))
		}
		s.lg.Info("shutdown http server")
	}
}

func (s *Server) Run() error {
	s.lg.Info("Start server", zap.String("addr", s.srv.Addr))
	return s.srv.ListenAndServe()
}

func NewRouter(ctx context.Context, auth authHandler, order orderHandler, balance balanceHandler, authService AuthService) *http.ServeMux {
	router := http.NewServeMux()

	apiUser := http.NewServeMux()

	// Публичные ручки
	apiUser.HandleFunc("POST /register", auth.Register)
	apiUser.HandleFunc("POST /login", auth.Login)

	// Приватные ручки
	privateMux := http.NewServeMux()
	privateMux.HandleFunc("POST /refresh", auth.Refresh)
	privateMux.HandleFunc("POST /logout", auth.Logout)
	privateMux.HandleFunc("POST /orders", order.Create)
	privateMux.HandleFunc("GET /orders", order.Get)
	privateMux.HandleFunc("GET /balance", balance.Get)
	privateMux.HandleFunc("POST /balance/withdraw", balance.Withdraw)
	privateMux.HandleFunc("GET /withdrawals", balance.History)

	// TODO: добавить auth middleware для проверки запросов к этим ручкам
	apiUser.Handle("/",
		middleware.Wrap(
			privateMux,
			middleware.WithAuth(authService),
		),
	)

	router.Handle(
		"/api/user/",
		middleware.Wrap(
			http.StripPrefix("/api/user", apiUser),
			middleware.WithLogger(zctx.From(ctx).Named("api logger")),
		),
	)

	return router
}
