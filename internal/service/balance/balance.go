package balance

import (
	"context"
	"strconv"

	"github.com/AA122AA/gomart.git/internal/db/query"
	"github.com/AA122AA/gomart.git/internal/domain"
	"github.com/AA122AA/gomart.git/internal/service/utils"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
)

type BalanceRepo interface {
	GetByUser(ctx context.Context, username string) (query.GetBalanceByUserNameRow, error)
	WithdrawTransaction(ctx context.Context, username string, oid int, amount float32) error
	GetBonusTransactions(ctx context.Context, username string) ([]query.GetBonusTransactionsByUserNameRow, error)
}

type BalanceService struct {
	repo BalanceRepo
	lg   *zap.Logger
}

func NewBalanceService(ctx context.Context, repo BalanceRepo) *BalanceService {
	return &BalanceService{
		repo: repo,
		lg:   zctx.From(ctx).Named("balance service"),
	}
}

func (bs *BalanceService) Get(ctx context.Context, username string) (*domain.BalanceJSON, error) {
	if username == "" {
		return nil, ErrEmptyUsername
	}

	balance, err := bs.repo.GetByUser(ctx, username)
	if err != nil {
		return nil, err
	}

	return domain.TransformBalanceToJSON(balance), nil
}

func (bs *BalanceService) Withdraw(ctx context.Context, username, oid string, amount float32) error {
	if username == "" || oid == "" || amount <= 0 {
		bs.lg.Error("empty data", zap.String("username", username), zap.String("oid", oid), zap.Float32("amount", amount))
		return ErrEmptyInputData
	}

	err := utils.IsLuna(oid)
	if err != nil {
		return err
	}

	noid, err := strconv.Atoi(oid)
	if err != nil {
		bs.lg.Error("can not parse oid", zap.String("oid", oid), zap.Error(err))
		return err
	}

	err = bs.repo.WithdrawTransaction(ctx, username, noid, amount)
	if err != nil {
		return err
	}

	return nil
}

func (bs *BalanceService) History(ctx context.Context, username string) ([]*domain.WithdrawJSON, error) {
	trans, err := bs.repo.GetBonusTransactions(ctx, username)
	if err != nil {
		return nil, err
	}

	res := make([]*domain.WithdrawJSON, 0, len(trans))
	for _, t := range trans {
		res = append(res, domain.TransformHistoryToJSON(t))
	}

	return res, nil
}
