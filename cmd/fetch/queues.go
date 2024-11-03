package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jadudm/eight/internal/env"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

// GLOBAL TO THE APP
// One pool of connections for River.
var dbPool *pgxpool.Pool

// The work client, doing the work of `fetch`
var fetchClient *river.Client[pgx.Tx]
var extractClient *river.Client[pgx.Tx]

// The enqueue client, for the API and others to enqueue work

func InitializeQueues() {
	ctx := context.Background()

	// Establsih the database
	database_url, err := env.Env.GetDatabaseUrl(env.WorkingDatabase)
	if err != nil {
		zap.L().Error("unable to get connection string; exiting",
			zap.String("database", env.WorkingDatabase),
		)
		os.Exit(1)
	}

	dbp, err := pgxpool.New(ctx, database_url)
	if err != nil {
		zap.L().Error("could not establish database pool; exiting",
			zap.String("database_url", database_url),
		)
		os.Exit(1)
	}
	// We want this pool for the workers.
	dbPool = dbp

	// Create a pool of workers
	workers := river.NewWorkers()
	river.AddWorker(workers, &FetchWorker{})

	// Grab the number of workers from the config.
	fetch_service, err := env.Env.GetUserService("fetch")
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	wC, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: int(fetch_service.GetParamInt64("workers"))},
		},
		Workers: workers,
	})

	if err != nil {
		zap.L().Error("could not establish worker pool")
		log.Println(err)
		os.Exit(1)
	}

	// Start the work clients
	if err := wC.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.")
		os.Exit(42)
	}

	// Insert-only client to `extract`
	eC, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{})
	if err != nil {
		zap.L().Error("could not establish insert-only client")
		os.Exit(1)
	}

	fetchClient = wC
	extractClient = eC
}
