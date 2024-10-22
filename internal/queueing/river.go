package queueing

import (
	"context"
	"log"
	"log/slog"
	"os"

	pgx "github.com/jackc/pgx/v5"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
	river "github.com/riverqueue/river"
	rpgxv5 "github.com/riverqueue/river/riverdriver/riverpgxv5"
	rivermigrate "github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivershared/util/slogutil"
)

type QueueNameType string

var QueueName = struct {
	Crawl string
}{
	Crawl: "crawl",
}

type River struct {
	Context   context.Context
	Pool      *pgxpool.Pool
	Client    *river.Client[pgx.Tx]
	QueueName string
}

func NewRiver(queue string) *River {
	r := River{}
	r.QueueName = queue
	return &r
}

func (r *River) Start() {

	// We need a context everywhere
	r.Context = context.Background()

	// Set up a pool
	pool, err := pgxpool.New(r.Context, os.Getenv("DATABASE_URL"))
	if err != nil {
		// handle error
	}
	r.Pool = pool

	// FIXME: This might close too soon.
	// defer r.Pool.Close()

	// Run the migrations, always.
	migrator, err := rivermigrate.New(rpgxv5.New(r.Pool), nil)
	if err != nil {
		log.Println("Could not create river migrator. Exiting.")
		log.Fatal(err)
	}
	_, err = migrator.Migrate(r.Context, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
	if err != nil {
		log.Println("Could not run river migrations. Exiting.")
		log.Fatal(err)
	}

	// Set up a client for queueing jobs

	client, err := river.NewClient(rpgxv5.New(r.Pool), &river.Config{
		Logger: slog.New(&slogutil.SlogMessageOnlyHandler{Level: slog.LevelWarn}),
	})
	if err != nil {
		log.Fatal(err)
	}
	r.Client = client
}

func (r *River) Insert(job river.JobArgs) {
	tx, err := r.Pool.Begin(r.Context)
	if err != nil {
		log.Println(err)
		log.Fatal("Could not create transaction")
	}
	_, err = r.Client.InsertTx(r.Context, tx, job, &river.InsertOpts{
		Queue: r.QueueName,
	})
	if err != nil {
		log.Println(err)
		log.Fatal("Could not insert with transaction")
	}
	tx.Commit(r.Context)
}
