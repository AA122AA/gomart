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
	UpdateBonusBalance(ctx context.Context, usernname string, balance, spent float32) error
	WriteBonusHistory(ctx context.Context, username string, oid int, amount float32) error
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
	balance, err := bs.repo.GetByUser(ctx, username)
	if err != nil {
		return nil, err
	}

	return domain.TransformBalanceToJSON(balance), nil
}

func (bs *BalanceService) Withdraw(ctx context.Context, username, oid string, amount float32) error {
	err := utils.IsLuna(oid)
	if err != nil {
		return err
	}

	bal, err := bs.repo.GetByUser(ctx, username)
	if err != nil {
		return err
	}

	if bal.CurrentBalance < amount {
		bs.lg.Error("no enought bonuses", zap.String("user", username), zap.Float32("want", amount), zap.Float32("have", bal.CurrentBalance))
		return NewErrNoMoney(nil)
	}

	newBal := bal.CurrentBalance - amount
	newSpent := bal.TotalBonusesSpent + amount

	err = bs.repo.UpdateBonusBalance(ctx, username, newBal, newSpent)
	if err != nil {
		return err
	}

	noid, err := strconv.Atoi(oid)
	if err != nil {
		return err
	}

	err = bs.repo.WriteBonusHistory(ctx, username, noid, amount)
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
