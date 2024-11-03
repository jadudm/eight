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

// GLOBAL TO THE APP
// One pool of connections for River.
// The work client, doing the work of `extract`
var extractClient *river.Client[pgx.Tx]
var extractPool *pgxpool.Pool

var walkClient *river.Client[pgx.Tx]
var walkPool *pgxpool.Pool

var packClient *river.Client[pgx.Tx]
var packPool *pgxpool.Pool

type ExtractWorker struct {
	river.WorkerDefaults[common.ExtractArgs]
}

func InitializeQueues() {
	var err error
	ctx, extractPool, workers := common.CommonQueueInit()
	_, walkPool, _ = common.CommonQueueInit()
	_, packPool, _ = common.CommonQueueInit()

	zap.L().Debug("initialized common queues")

	river.AddWorker(workers, &ExtractWorker{})

	// Grab the number of workers from the config.
	extract_service, err := env.Env.GetUserService("extract")
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	extractClient, err = river.NewClient(riverpgxv5.New(extractPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			"extract": {MaxWorkers: int(extract_service.GetParamInt64("workers"))},
		},
		Workers: workers,
	})

	if err != nil {
		zap.L().Error("could not establish worker pool")
		log.Println(err)
		os.Exit(1)
	}

	// write-only clients for posting jobs
	walkClient, err = river.NewClient(riverpgxv5.New(walkPool), &river.Config{})
	if err != nil {
		zap.L().Error("could not start insert-only walk client")
	}

	packClient, err = river.NewClient(riverpgxv5.New(packPool), &river.Config{})
	if err != nil {
		zap.L().Error("could not start insert-only pack client")
	}

	// Start the work clients
	if err := extractClient.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.")
		os.Exit(42)
	}
}
