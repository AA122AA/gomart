package accrual

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"testing"

	"github.com/AA122AA/gomart.git/internal/config"
	"github.com/AA122AA/gomart.git/internal/db"
	"github.com/AA122AA/gomart.git/internal/db/query"
	"github.com/AA122AA/gomart.git/internal/repoistory"
)

func TestUpdateOrders(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	var wg sync.WaitGroup

	cfg := &config.Config{
		AccrualHost: "http://localhost:8080",
	}
	DatabaseDSN := "postgresql://gomart:StrongPass123!@localhost:5432/gomart?sslmode=disable"

	database := db.New(ctx, DatabaseDSN)
	queries := query.New(database.DB())

	repo := repoistory.NewAccrualRepo(ctx, queries, database)
	ac := NewAccrualClient(ctx, repo, cfg)

	orderCh := ac.getOrders(ctx, &wg)
	in := ac.getAccruals(ctx, &wg, orderCh)
	wg.Add(1)
	go ac.updateOrders(ctx, &wg, in)

	fmt.Printf("waiting\n")
	wg.Wait()

}
