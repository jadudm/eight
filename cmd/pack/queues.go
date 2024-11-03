package main

import (
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

var packPool *pgxpool.Pool
var packClient *river.Client[pgx.Tx]
var extractPool *pgxpool.Pool
var extractClient *river.Client[pgx.Tx]
var servePool *pgxpool.Pool
var serveClient *river.Client[pgx.Tx]

type PackWorker struct {
	river.WorkerDefaults[common.PackArgs]
}

func InitializeQueues() {
	ctx, pP, workers := common.CommonQueueInit()
	_, eP, _ := common.CommonQueueInit()
	_, sP, _ := common.CommonQueueInit()
	packPool = pP
	extractPool = eP
	servePool = sP

	// Essentially adds a worker "type" to the work engine.
	river.AddWorker(workers, &PackWorker{})

	// Grab the number of workers from the config.
	fetch_service, err := env.Env.GetUserService("fetch")
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	packClient, err = river.NewClient(riverpgxv5.New(packPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			"pack": {MaxWorkers: int(fetch_service.GetParamInt64("workers"))},
		},
		Workers: workers,
	})

	if err != nil {
		zap.L().Error("could not establish worker pool")
		log.Println(err)
		os.Exit(1)
	}

	// Insert-only client to `extract`
	extractClient, err = river.NewClient(riverpgxv5.New(extractPool), &river.Config{})
	if err != nil {
		zap.L().Error("could not establish insert-only client")
		os.Exit(1)
	}

	serveClient, err = river.NewClient(riverpgxv5.New(servePool), &river.Config{})
	if err != nil {
		zap.L().Error("could not establish insert-only client")
		os.Exit(1)
	}

	// Start the work clients
	if err := packClient.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.")
		os.Exit(42)
	}
}
