package common

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jadudm/eight/internal/env"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

func CommonQueueInit() (context.Context, *pgxpool.Pool, *river.Workers) {
	var err error
	ctx := context.Background()

	// Establsih the database
	database_url, err := env.Env.GetDatabaseUrl(env.WorkingDatabase)
	if err != nil {
		zap.L().Error("unable to get connection string; exiting",
			zap.String("database", env.WorkingDatabase),
		)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, database_url)
	if err != nil {
		zap.L().Error("could not establish database pool; exiting",
			zap.String("database_url", database_url),
		)
		os.Exit(1)
	}

	// Create a pool of workers
	workers := river.NewWorkers()
	return ctx, pool, workers

}

func CtxTx(pool *pgxpool.Pool) (context.Context, pgx.Tx) {
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		zap.L().Panic("cannot init tx from pool")
	}
	//defer tx.Rollback(ctx)

	return ctx, tx
}