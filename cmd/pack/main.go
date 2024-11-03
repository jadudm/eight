package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/sqlite"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

var ch_finalize chan *sqlite.PackTable

// We get pings on domains as they go through
// When the timer fires, we queue that domain to the finalize queue.
func FinalizeTimer(in <-chan *sqlite.PackTable) {
	//FIXME: This should be a config parameter
	TIMEOUT_DURATION := time.Duration(10 * time.Second)

	clocks := make(map[string]time.Time)
	tables := make(map[string]*sqlite.PackTable)

	// https://dev.to/milktea02/misunderstanding-go-timers-and-channels-1jal
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
			zap.L().Debug("finalize timeout")
			for sqlite_filename, clock := range clocks {
				if time.Since(clock) > TIMEOUT_DURATION {
					//prw.EnqueueClient()
					// FIXME: Just send it to S3 for now.
					// This is still a bit of an MVP.

					zap.L().Debug("finalize streaming",
						zap.String("sqlite_filename", sqlite_filename))

					tables[sqlite_filename].PrepForNetwork()

					err := serveStorage.StoreFile(sqlite_filename, sqlite_filename)
					if err != nil {
						log.Println("PACK could not store to file", sqlite_filename)
						log.Fatal(err)
					}

					// Enqueue serve
					zap.L().Debug("inserting serve job")
					ctx, tx := common.CtxTx(extractPool)
					serveClient.InsertTx(ctx, tx, common.ExtractArgs{
						Key: sqlite_filename,
					}, &river.InsertOpts{Queue: "serve"})
					if err := tx.Commit(ctx); err != nil {
						tx.Rollback(ctx)
						zap.L().Panic("cannot commit insert tx",
							zap.String("key", sqlite_filename))
					}

					delete(clocks, sqlite_filename)
					delete(tables, sqlite_filename)

				}
			}
		}
		zap.L().Debug("finalize reset")
		timeout.Reset(TIMEOUT_DURATION)
	}
}

func main() {
	env.InitGlobalEnv()
	InitializeQueues()
	InitializeStorage()
	engine := common.InitializeAPI()
	log.Println("environment initialized")

	ch_finalize = make(chan *sqlite.PackTable)
	go FinalizeTimer(ch_finalize)

	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, engine)
}
