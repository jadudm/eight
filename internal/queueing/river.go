package queueing

import (
	"context"
	"log"
	"log/slog"
	"time"

	pgx "github.com/jackc/pgx/v5"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
	"github.com/jadudm/eight/internal/env"
	river "github.com/riverqueue/river"
	rpgxv5 "github.com/riverqueue/river/riverdriver/riverpgxv5"
	rivermigrate "github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivershared/util/slogutil"
)

// type QueueNameType string

// var QueueName = struct {
// 	Crawl string
// }{
// 	Crawl: "crawl",
// }

type River struct {
	Context   context.Context
	Pool      *pgxpool.Pool
	Client    *river.Client[pgx.Tx]
	Logger    *slog.Logger
	QueueName string
}

func (r *River) initialize() {

	// We need a context everywhere
	r.Context = context.Background()

	// Set up a pool
	connection_string, err := env.Env.GetDatabaseUrl(env.WorkingDatabase)

	if err != nil {
		log.Println("RIVER cannot find connection string for", env.WorkingDatabase)
		log.Fatal(err)
	}
	pool, err := pgxpool.New(r.Context, connection_string)
	if err != nil {
		// handle error
	}
	r.Pool = pool

	// FIXME: This might close too soon.
	// defer r.Pool.Close()

	// Run the migrations, always.
	migrator, err := rivermigrate.New(rpgxv5.New(r.Pool), nil)
	if err != nil {
		log.Println("RIVER Could not create river migrator. Exiting.")
		log.Fatal(err)
	}
	_, err = migrator.Migrate(r.Context, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
	if err != nil {
		log.Println("RIVER Could not run river migrations. Exiting.")
		log.Fatal(err)
	}

	r.Logger = slog.New(&slogutil.SlogMessageOnlyHandler{Level: slog.LevelWarn})
}

func NewRiver() *River {
	r := River{}
	r.initialize()
	return &r
}

func QueueingClient(r *River, job river.JobArgs) *River {
	r.QueueName = job.Kind()

	client, err := river.NewClient(rpgxv5.New(r.Pool), &river.Config{
		Logger: r.Logger,
	})

	if err != nil {
		log.Println("RIVER Could not initialize river client")
		log.Fatal(err)
	}

	r.Client = client
	return r
}

func WorkingClient[T river.JobArgs, U river.Worker[T]](r *River, job T, worker river.Worker[T]) *River {
	r.QueueName = job.Kind()

	workers := river.NewWorkers()
	river.AddWorker(workers, worker)

	rc, err := river.NewClient(rpgxv5.New(r.Pool), &river.Config{
		Logger: r.Logger,
		Queues: map[string]river.QueueConfig{
			job.Kind(): {MaxWorkers: 10},
		},
		Workers: workers,
		// Explore these parameters. The rescue lets us pick up jobs
		// that were part-way done (say, in case of a crash).
		FetchCooldown:        time.Duration(1 * time.Second),
		FetchPollInterval:    time.Duration(2 * time.Second),
		JobTimeout:           time.Duration(10 * time.Second),
		RescueStuckJobsAfter: time.Duration(30 * time.Second),
	})
	if err != nil {
		panic(err)
	}
	r.Client = rc
	return r
}

func (r *River) InsertTx(job river.JobArgs) {
	tx, err := r.Pool.Begin(r.Context)
	if err != nil {
		log.Println(err)
		log.Fatal("Could not create transaction")
	}
	// _, err = r.Client.InsertTx(r.Context, tx, job, &river.InsertOpts{Queue: r.QueueName})
	_, err = r.Client.InsertTx(r.Context, tx, job, nil)
	if err != nil {
		log.Println(err)
		log.Fatal("Could not insert with transaction")
	}
	tx.Commit(r.Context)
}

func (r *River) Insert(job river.JobArgs, queuename string) {
	// _, err = r.Client.InsertTx(r.Context, tx, job, &river.InsertOpts{Queue: r.QueueName})
	_, err := r.Client.Insert(r.Context, job, &river.InsertOpts{Queue: queuename})
	if err != nil {
		log.Println(err)
		log.Fatal("Could not insert with transaction")
	}
}
