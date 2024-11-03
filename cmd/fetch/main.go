package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/pkg/fetch"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

// GLOBAL TO THE APP
// One pool of connections for River.
var dbPool *pgxpool.Pool

// The work client, doing the work of `fetch`
var workClient river.Client[pgx.Tx]

// The enqueue client, for the API and others to enqueue work

func InitializeRiver() {
	ctx := context.Background()

	// Establsih the database
	database_url, err := env.Env.GetDatabaseUrl(env.WorkingDatabase)
	zap.L().Error("could not get database URL",
		zap.String("database", env.WorkingDatabase),
	)

	dbPool, err := pgxpool.New(ctx, database_url)
	if err != nil {
		zap.L().Error("could not establish database pool; exiting",
			zap.String("database_url", database_url),
		)
		os.Exit(1)
	}

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

	workClient = wC
}

func main() {
	env.InitGlobalEnv()
	log.Println("environment initialized")

	ch := make(chan *fetch.FetchRequest)

	r := api.BaseMux()
	extended_api := FetchApi(r, ch)

	go fetch.Fetch(ch)

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))

	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, extended_api)

}
