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
