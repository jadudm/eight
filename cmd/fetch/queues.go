package main

import (
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	common "github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

// GLOBAL TO THE APP
// One pool of connections for River.

// The work client, doing the work of `fetch`
var dbPool *pgxpool.Pool
var fetchClient *river.Client[pgx.Tx]
var extractClient *river.Client[pgx.Tx]

type FetchWorker struct {
	river.WorkerDefaults[common.FetchArgs]
}

func InitializeQueues() {
	ctx, pool, workers := common.CommonQueueInit()
	dbPool = pool

	// Essentially adds a worker "type" to the work engine.
	river.AddWorker(workers, &FetchWorker{})

	// Grab the number of workers from the config.
	fetch_service, err := env.Env.GetUserService("fetch")
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	fetchClient, err = river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			"fetch": {MaxWorkers: int(fetch_service.GetParamInt64("workers"))},
		},
		Workers: workers,
	})

	if err != nil {
		zap.L().Error("could not establish worker pool")
		log.Println(err)
		os.Exit(1)
	}

	// Insert-only client to `extract`
	extractClient, err = river.NewClient(riverpgxv5.New(dbPool), &river.Config{})
	if err != nil {
		zap.L().Error("could not establish insert-only client")
		os.Exit(1)
	}

	// Start the work clients
	if err := fetchClient.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.")
		os.Exit(42)
	}
}
