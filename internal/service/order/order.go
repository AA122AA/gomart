package order

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/AA122AA/gomart.git/internal/constants"
	"github.com/AA122AA/gomart.git/internal/db/query"
	"github.com/AA122AA/gomart.git/internal/domain"
	"github.com/AA122AA/gomart.git/internal/service/utils"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type OrderRepo interface {
	Create(ctx context.Context, OID int, username, status string) error
	Get(ctx context.Context, ouid int) (query.GetOrderByOIDRow, error)
	GetAll(ctx context.Context) ([]query.GetAllRow, error)
	GetByUser(ctx context.Context, username string) ([]query.GetOrdersByUserNameRow, error)
}

type OrderService struct {
	repo OrderRepo
	lg   *zap.Logger
}

func NewOrderService(ctx context.Context, repo OrderRepo) *OrderService {
	return &OrderService{
		repo: repo,
		lg:   zctx.From(ctx).Named("order service"),
	}
}

func (osrv *OrderService) Get(ctx context.Context, username string) ([]*domain.OrderJSON, error) {
	orders, err := osrv.repo.GetByUser(ctx, username)
	if err != nil {
		return nil, err
	}

	res := make([]*domain.OrderJSON, 0, len(orders))
	for _, o := range orders {
		res = append(res, domain.TransformOrderToJSON(o))
	}

	return res, nil
}

func (osrv *OrderService) GetAll(ctx context.Context) ([]query.GetAllRow, error) {
	return osrv.repo.GetAll(ctx)
}

func (osrv *OrderService) Create(ctx context.Context, OID, username string) error {
	noid, err := strconv.Atoi(OID)
	if err != nil {
		return NewErrBadOrderID(err)
	}

	orderInfo, err := osrv.repo.Get(ctx, noid)
	if err == nil {
		if orderInfo.Username != username {
			return NewErrWrongUser(nil)
		}
		return NewErrOrderExists(nil)
	}

	if strings.Contains(err.Error(), "no rows in result set") {
		err = utils.IsLuna(OID)
		if err != nil {
			return err
		}

		err = osrv.repo.Create(ctx, noid, username, constants.New)
		if err != nil {
			return fmt.Errorf("error while writing to db: %w", err)
		}

		return nil
	}

	return err
}
