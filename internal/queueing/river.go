package queueing

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jadudm/eight/internal/env"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"go.uber.org/zap"
)

func InitializeRiverQueues() {
	// Set up a pool
	connection_string, err := env.Env.GetDatabaseUrl(env.WorkingDatabase)
	if err != nil {
		zap.L().Fatal("cannot find db connection string",
			zap.String("database", env.WorkingDatabase))
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connection_string)
	if err != nil {
		zap.L().Fatal("cannot create database pool for migrations")
	}
	defer pool.Close()

	// Run the migrations, always.
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		zap.L().Info("could not create a river migrator")
	}
	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
	if err != nil {
		zap.L().Info("could not run the river migrator")
	}
}
