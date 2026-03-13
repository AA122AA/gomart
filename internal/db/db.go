package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AA122AA/gomart.git/db/schema"
	"github.com/go-faster/sdk/zctx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type DB struct {
	pool *pgxpool.Pool

	lg *zap.Logger
}

func New(ctx context.Context, dsn string) *DB {
	lg := zctx.From(ctx).Named("Database")
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		lg.Fatal("cannot open db", zap.Error(err))
	}

	return &DB{
		pool: db,
		lg:   lg,
	}
}

func (d *DB) BeginTX(ctx context.Context) (pgx.Tx, error) {
	return d.pool.Begin(ctx)
}

func (d *DB) DB() *pgxpool.Pool {
	return d.pool
}

func (d *DB) Migrate(ctx context.Context) error {
	goose.SetBaseFS(schema.Migrations)

	const cmd = "up"
	db, err := sql.Open("pgx", d.pool.Config().ConnString())
	if err != nil {
		d.lg.Error("cannot connect to db in migrations", zap.Error(err))
		return fmt.Errorf("cannot run migrations: %w", err)
	}
	defer db.Close()

	err = goose.RunContext(ctx, cmd, db, ".")
	if err != nil {
		d.lg.Error("cannot run migrations", zap.Error(err))
		return fmt.Errorf("cannot run migrations: %w", err)
	}

	return nil
}

func (d *DB) Ping(ctx context.Context) error {
	d.lg.Debug("ping db")
	return d.pool.Ping(ctx)
}

func (d *DB) Close() {
	d.lg.Info("closing db connection")
	d.pool.Close()

}
