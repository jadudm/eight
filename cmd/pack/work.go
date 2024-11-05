package main

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/sqlite"
	"github.com/jadudm/eight/internal/sqlite/schemas"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

var databases sync.Map

var ch_finalize chan *sqlite.PackTable

// We get pings on domains as they go through
// When the timer fires, we queue that domain to the finalize queue.
func FinalizeTimer(in <-chan *sqlite.PackTable) {
	clocks := make(map[string]time.Time)
	tables := make(map[string]*sqlite.PackTable)

	// https://dev.to/milktea02/misunderstanding-go-timers-and-channels-1jal
	s, _ := env.Env.GetUserService("pack")
	TIMEOUT_DURATION := time.Duration(s.GetParamInt64("packing_timeout_seconds")) * time.Second
	zap.L().Debug("finalize starting timer")
	timeout := time.NewTimer(TIMEOUT_DURATION)

	for {
		select {
		case pt := <-in:
			// When we get a domain, we should indicate that we
			// saw it just now.
			clocks[pt.Filename] = time.Now()
			tables[pt.Filename] = pt

		case <-timeout.C:
			// Every <timeout> seconds, we'll see if anyone has a clock that is greater,
			// which will mean nothing has come through recently.
			//zap.L().Debug("finalize timeout")
			for sqlite_filename, clock := range clocks {
				if time.Since(clock) > TIMEOUT_DURATION {
					zap.L().Info("packing to sqlite",
						zap.String("sqlite_filename", sqlite_filename))

					tables[sqlite_filename].PrepForNetwork()

					err := serveStorage.StoreFile(sqlite_filename, sqlite_filename)
					if err != nil {
						log.Println("PACK could not store to file", sqlite_filename)
						log.Fatal(err)
					}

					// Enqueue serve
					zap.L().Debug("inserting serve job")
					ctx, tx := common.CtxTx(servePool)
					serveClient.InsertTx(ctx, tx, common.ServeArgs{
						Filename: sqlite_filename,
					}, &river.InsertOpts{Queue: "serve"})
					if err := tx.Commit(ctx); err != nil {
						tx.Rollback(ctx)
						zap.L().Panic("cannot commit insert tx",
							zap.String("filename", sqlite_filename))
					}

					delete(clocks, sqlite_filename)
					delete(tables, sqlite_filename)

				}
			}
		}
		//zap.L().Debug("finalize reset")
		timeout.Reset(TIMEOUT_DURATION)
	}
}

func (w *PackWorker) Work(ctx context.Context, job *river.Job[common.PackArgs]) error {
	zap.L().Debug("packing")

	obj, err := extractStorage.Get(job.Args.Key)
	if err != nil {
		zap.L().Fatal("cannot get object from S3",
			zap.String("key", job.Args.Key),
		)
	}
	JSON := obj.GetJson()

	host := JSON["host"]

	if _, ok := databases.Load(host); !ok {
		table, err := sqlite.CreatePackTable(sqlite.SqliteFilename(host), JSON)
		if err != nil {
			log.Println("Could not create pack table for", host)
			log.Fatal(err)
		}
		databases.Store(host, table)
	}

	if _pt, ok := databases.Load(host); ok {
		pt := _pt.(*sqlite.PackTable)
		_, err := pt.Queries.CreateSiteEntry(pt.Context, schemas.CreateSiteEntryParams{
			Host: JSON["host"],
			Path: JSON["path"],
			Text: JSON["content"],
		})
		if err != nil {
			log.Println("Insert into site entry table failed")
			log.Fatal(err)
		}
		zap.L().Info("packed entry",
			zap.String("database", host),
			zap.String("path", JSON["path"]),
			zap.Int("length", len(JSON["content"])))

		ch_finalize <- pt
	}

	// Agressively keep memory clear.
	// GC after packing every message.
	runtime.GC()
	return nil
}
