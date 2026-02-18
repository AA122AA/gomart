package repoistory

import (
	"context"
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

func (ur *UserRepo) Create(ctx context.Context, username, passwordHash string) error {
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

	return ur.queries.InsertUser(ctx, params)
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

func (ur *UserRepo) InvalidedRefreshToken(ctx context.Context, username string) error {
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
		lg:      zctx.From(ctx).Named("user repo"),
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
