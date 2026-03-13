package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/AA122AA/gomart.git/internal/config"
	"github.com/AA122AA/gomart.git/internal/db/query"
	"github.com/AA122AA/gomart.git/internal/domain"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type AccrualRepo interface {
	GetAll(ctx context.Context) ([]query.GetAllOrdersForAccrualRow, error)
	UpdateAccruals(ctx context.Context, accruals map[string]*domain.AccrualResponseJSON) error
}

type AccrualClient struct {
	url  string
	repo AccrualRepo

	client *http.Client
	lg     *zap.Logger
}

func NewAccrualClient(ctx context.Context, repo AccrualRepo, cfg *config.Config) *AccrualClient {
	return &AccrualClient{
		url:    cfg.AccrualHost,
		repo:   repo,
		client: &http.Client{},
		lg:     zctx.From(ctx).Named("accrual client"),
	}
}

func (ac *AccrualClient) Run(ctx context.Context, wg *sync.WaitGroup) {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	orderCh := ac.getOrders(ctx, wg)
	accrualCh := ac.getAccruals(ctx, wg, orderCh)
	wg.Add(1)
	go ac.updateOrders(ctx, wg, accrualCh)
}

func (ac *AccrualClient) updateOrders(ctx context.Context, wg *sync.WaitGroup, in <-chan map[string]*domain.AccrualResponseJSON) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			ac.lg.Info("got cancellation, returning ")
			return
		case accrual := <-in:
			err := ac.repo.UpdateAccruals(ctx, accrual)
			if err != nil {
				fmt.Printf("error - %v", err)
				return
			}

		}
	}
}

func (ac *AccrualClient) getAccruals(ctx context.Context, wg *sync.WaitGroup, in <-chan int) chan map[string]*domain.AccrualResponseJSON {
	ticker := time.NewTicker(5 * time.Second)
	accrualCh := make(chan map[string]*domain.AccrualResponseJSON, 1)
	mu := &sync.RWMutex{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ticker.Stop()
		res := make(map[string]*domain.AccrualResponseJSON)
		for {
			select {
			case <-ctx.Done():
				ac.lg.Info("got cancellation, returning ")
				close(accrualCh)
				return
			case order := <-in:
				resp, err := ac.makeRequest(order)
				if err != nil {
					ac.lg.Error("error while making request", zap.Error(err))
					continue
				}
				mu.Lock()
				res[resp.Order] = resp
				mu.Unlock()
			case <-ticker.C:
				mu.RLock()
				accrualCh <- maps.Clone(res)
				mu.RUnlock()
			}
		}
	}()

	return accrualCh
}

func (ac *AccrualClient) getOrders(ctx context.Context, wg *sync.WaitGroup) chan int {
	orderCh := make(chan int, 40)
	ticker := time.NewTicker(10 * time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				ac.lg.Info("got cancellation, returning ")
				close(orderCh)
				return
			case <-ticker.C:
				//TODO: сделать чтоб доставал только в статусе NEW+
				res, err := ac.repo.GetAll(ctx)
				if err != nil {
					ac.lg.Error("got error while getting orders", zap.Error(err))
					continue
				}

				for _, o := range res {
					orderCh <- int(o.Oid)
				}
			}
		}
	}()

	return orderCh
}

func (ac *AccrualClient) makeRequest(oid int) (*domain.AccrualResponseJSON, error) {
	fmt.Printf("in makeRequest, oid - %v\n", oid)
	u, err := url.Parse(ac.url)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "/api/orders/", fmt.Sprintf("%v", oid))
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := ac.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ac.lg.Error("status not OK", zap.Int("status code", resp.StatusCode), zap.String("status", resp.Status))
		return nil, err
	}

	var r domain.AccrualResponseJSON
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	return &r, nil

}
