package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/AA122AA/gomart.git/internal/config"
	"github.com/AA122AA/gomart.git/internal/db"
	"github.com/AA122AA/gomart.git/internal/db/query"
	authhandler "github.com/AA122AA/gomart.git/internal/handler/auth"
	"github.com/AA122AA/gomart.git/internal/handler/balance"
	"github.com/AA122AA/gomart.git/internal/handler/order"
	"github.com/AA122AA/gomart.git/internal/repoistory"
	"github.com/AA122AA/gomart.git/internal/server"
	authservice "github.com/AA122AA/gomart.git/internal/service/auth"
	"github.com/AA122AA/gomart.git/internal/zapcfg"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

func main() {
	lg, err := zapcfg.New().Build()
	if err != nil {
		log.Fatalf("got error while creating logger - %v", err)
	}

	flush := func() { lg.Sync() }
	defer flush()

	defer func() {
		if r := recover(); r != nil {
			lg.Fatal("Panic recovering", zap.Any("panic", r))
			os.Exit(2)
		}
	}()

	ctx, cancel := signal.NotifyContext(zctx.Base(context.Background(), lg), os.Interrupt)
	defer cancel()

	cfg := &config.Config{}
	err = cfg.ParseConfig()
	if err != nil {
		lg.Fatal("error while parsing config", zap.Error(err))
	}

	lg.Debug(
		"config",
		zap.String("address", cfg.HostAddr),
	)

	// Init DataBase and migrate
	database := db.New(ctx, cfg.DatabaseDSN)
	err = database.Migrate(ctx)
	if err != nil {
		lg.Fatal("cannot make migration", zap.Error(err))
	}

	// Init repo
	queries := query.New(database.DB())
	authRepo := repoistory.NewUserRepo(ctx, queries, database)

	var wg sync.WaitGroup

	// Init services
	authService := authservice.NewUserService(ctx, authRepo, cfg)

	// Init handlers
	authHandler := authhandler.NewAuthHandler(ctx, authService)
	orderHandler := order.NewOrderHandler(ctx)
	balanceHandler := balance.NewBalanceHandler(ctx)

	// Init router
	router := server.NewRouter(ctx, authHandler, orderHandler, balanceHandler, authService)

	// Init server
	srv := server.NewServer(ctx, router, cfg)

	wg.Add(1)
	go srv.OnShutDown(ctx, &wg)

	srv.Run()

	wg.Wait()
}
