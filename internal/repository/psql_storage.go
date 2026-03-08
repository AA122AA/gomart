package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/AA122AA/gomart.git/internal/db"
	"github.com/AA122AA/gomart.git/internal/db/query"
	"github.com/AA122AA/gomart.git/internal/domain"
	"github.com/go-faster/sdk/zctx"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// --- User Repository --- \\
type UserRepo struct {
	db      *db.DB
	queries *query.Queries
	lg      *zap.Logger
}

func NewUserRepo(ctx context.Context, queries *query.Queries, db *db.DB) *UserRepo {
	return &UserRepo{
		db:      db,
		queries: queries,
		lg:      zctx.From(ctx).Named("user repo"),
	}
}

func (ur *UserRepo) CreateUserAndBonusAcc(ctx context.Context, username, passwordHash string, balance, spent float32) error {
	tx, err := ur.db.BeginTX(ctx)
	if err != nil {
		ur.lg.Error("cannot begin TX", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)

	q := ur.queries.WithTx(tx)

	now := pgtype.Timestamptz{
		Time:  time.Now(),
		Valid: true,
	}

	params := query.InsertUserParams{
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = q.InsertUser(ctx, params)
	if err != nil {
		ur.lg.Error("failed to insert user", zap.Error(err))
		return err
	}

	baParams := query.CreateBonusAccountParams{
		Username:          username,
		CurrentBalance:    balance,
		TotalBonusesSpent: spent,
	}
	err = q.CreateBonusAccount(ctx, baParams)
	if err != nil {
		ur.lg.Error("failed to create bonus acc", zap.Error(err))
		return err
	}

	return tx.Commit(ctx)
}

func (ur *UserRepo) CreateBonusAcc(ctx context.Context, username string, balance, spent float32) error {
	params := query.CreateBonusAccountParams{
		Username:          username,
		CurrentBalance:    balance,
		TotalBonusesSpent: spent,
	}

	return ur.queries.CreateBonusAccount(ctx, params)
}

func (ur *UserRepo) GetPasswordHash(ctx context.Context, username string) (string, error) {
	return ur.queries.GetUserPassword(ctx, username)
}

func (ur *UserRepo) CreateRefreshToken(ctx context.Context, tokenID, username string, isValid bool) error {

	params := query.InsertRefreshTokenParams{
		Username: username,
		TokenID:  tokenID,
		IsValid:  true,
	}

	return ur.queries.InsertRefreshToken(ctx, params)
}

func (ur *UserRepo) GetRefreshTokenByUserName(ctx context.Context, username string) (*domain.GetRefreshTokenID, error) {
	res, err := ur.queries.GetRefreshTokenByUserName(ctx, username)
	if err != nil {
		return nil, err
	}

	return &domain.GetRefreshTokenID{
		TokenID: res.TokenID,
		IsValid: res.IsValid,
	}, nil
}

func (ur *UserRepo) UpdateRefreshToken(ctx context.Context, tokenID, username string, isValid bool) error {

	params := query.UpdateRefreshTokenParams{
		TokenID:  tokenID,
		IsValid:  isValid,
		Username: username,
	}

	return ur.queries.UpdateRefreshToken(ctx, params)
}

func (ur *UserRepo) InvalidateRefreshToken(ctx context.Context, username string) error {
	params := query.UpdateRefreshTokenIsValidParams{
		Username: username,
		IsValid:  false,
	}

	return ur.queries.UpdateRefreshTokenIsValid(ctx, params)
}

// --- Order Repository --- \\
type OrderRepo struct {
	db      *db.DB
	queries *query.Queries
	lg      *zap.Logger
}

func NewOrderRepo(ctx context.Context, queries *query.Queries, db *db.DB) *OrderRepo {
	return &OrderRepo{
		db:      db,
		queries: queries,
		lg:      zctx.From(ctx).Named("order repo"),
	}
}

func (or *OrderRepo) Create(ctx context.Context, OID int, username, status string) error {
	now := pgtype.Timestamptz{
		Time:  time.Now(),
		Valid: true,
	}
	params := query.CreateOrderParams{
		Oid:        int64(OID),
		Username:   username,
		Status:     status,
		UploadedAt: now,
		UpdatedAt:  now,
	}

	return or.queries.CreateOrder(ctx, params)
}

func (or *OrderRepo) Get(ctx context.Context, oid int) (query.GetOrderByOIDRow, error) {
	return or.queries.GetOrderByOID(ctx, int64(oid))
}

func (or *OrderRepo) GetAll(ctx context.Context) ([]query.GetAllRow, error) {
	return or.queries.GetAll(ctx)
}

func (or *OrderRepo) GetByUser(ctx context.Context, username string) ([]query.GetOrdersByUserNameRow, error) {
	return or.queries.GetOrdersByUserName(ctx, username)
}

func (or *OrderRepo) UpdateOrderStatus(ctx context.Context, accruals map[string]*domain.AccrualResponseJSON) error {
	tx, err := or.db.BeginTX(ctx)
	if err != nil {
		or.lg.Error("cannot begin TX", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)

	q := or.queries.WithTx(tx)
	for _, v := range accruals {
		noid, err := strconv.Atoi(v.Order)
		if err != nil {
			return err
		}
		params := query.UpdateOrderStatusParams{
			Status: v.Status,
			Accrual: pgtype.Float4{
				Float32: v.Accrual,
				Valid:   true,
			},
			Oid: int64(noid),
		}

		err = q.UpdateOrderStatus(ctx, params)
		if err != nil {
			return err
		}
		fmt.Printf("updated order - %v\n", v.Order)

	}

	fmt.Printf("done updating, gonna commit\n")

	return tx.Commit(ctx)
}

// --- Balance Repository --- \\
type BalanceRepo struct {
	db      *db.DB
	queries *query.Queries
	lg      *zap.Logger
}

func NewBalanceRepo(ctx context.Context, queries *query.Queries, db *db.DB) *BalanceRepo {
	return &BalanceRepo{
		db:      db,
		queries: queries,
		lg:      zctx.From(ctx).Named("balance repo"),
	}
}

func (br *BalanceRepo) GetByUser(ctx context.Context, username string) (query.GetBalanceByUserNameRow, error) {
	return br.queries.GetBalanceByUserName(ctx, username)
}

func (br *BalanceRepo) UpdateBonusBalance(ctx context.Context, username string, balance, spent float32) error {
	params := query.UpdateBalanceParams{
		CurrentBalance:    balance,
		TotalBonusesSpent: spent,
		Username:          username,
	}

	return br.queries.UpdateBalance(ctx, params)
}

func (br *BalanceRepo) WriteBonusHistory(ctx context.Context, username string, oid int, amount float32) error {
	now := pgtype.Timestamptz{
		Time:  time.Now(),
		Valid: true,
	}
	params := query.InsertBonusTransactionParams{
		Username:        username,
		Oid:             int64(oid),
		BonusesWithdraw: amount,
		ProcessedAt:     now,
	}

	return br.queries.InsertBonusTransaction(ctx, params)
}

func (br *BalanceRepo) GetBonusTransactions(ctx context.Context, username string) ([]query.GetBonusTransactionsByUserNameRow, error) {
	return br.queries.GetBonusTransactionsByUserName(ctx, username)
}

// --- Accrual Repository --- \\
type AccrualRepo struct {
	db      *db.DB
	queries *query.Queries
	lg      *zap.Logger
}

func NewAccrualRepo(ctx context.Context, queries *query.Queries, db *db.DB) *AccrualRepo {
	return &AccrualRepo{
		db:      db,
		queries: queries,
		lg:      zctx.From(ctx).Named("accrual repo"),
	}
}

func (ar *AccrualRepo) GetAll(ctx context.Context) ([]query.GetAllOrdersForAccrualRow, error) {
	return ar.queries.GetAllOrdersForAccrual(ctx)
}

func (ar *AccrualRepo) UpdateAccruals(ctx context.Context, accruals map[string]*domain.AccrualResponseJSON) error {
	tx, err := ar.db.BeginTX(ctx)
	if err != nil {
		ar.lg.Error("cannot begin TX", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)

	q := ar.queries.WithTx(tx)
	for _, v := range accruals {
		noid, err := strconv.Atoi(v.Order)
		if err != nil {
			return err
		}
		params := query.UpdateOrderStatusParams{
			Status: v.Status,
			Accrual: pgtype.Float4{
				Float32: v.Accrual,
				Valid:   true,
			},
			Oid: int64(noid),
		}

		err = q.UpdateOrderStatus(ctx, params)
		if err != nil {
			return err
		}
		fmt.Printf("updated order - %v\n", v.Order)

		res, err := q.GetBalanceByOrderID(ctx, int64(noid))
		if err != nil {
			return err
		}

		bParams := query.UpdateBalanceParams{
			CurrentBalance:    res.CurrentBalance + v.Accrual,
			TotalBonusesSpent: res.TotalBonusesSpent,
			Username:          res.Username,
		}

		err = q.UpdateBalance(ctx, bParams)
		if err != nil {
			return err
		}
	}

	fmt.Printf("done updating, gonna commit\n")

	return tx.Commit(ctx)
}
