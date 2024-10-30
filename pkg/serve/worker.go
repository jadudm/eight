package serve

import (
	"context"
	"log"

	"search.eight/internal/sqlite"
)

// The worker just grabs things off the queue and
// spits them out the channel. The Crawl proc then
// does the work of processing it.
func (srw *ServeRequestWorker) Work(
	ctx context.Context,
	job *ServeRequestJob,
) error {

	// Our requests will be for packed DBs.
	// So, those should be fetched into the local system, and
	// the API should know that we can now serve the domain.
	JSON, err := srw.FetchStorage.Get(job.Args.Key)
	if err != nil {
		log.Fatal(err)
	}

	sqlite_filename := sqlite.SqliteFilename(JSON["host"])
	// Writes to the local filesystem.
	srw.ServeStorage.GetObject(sqlite_filename, sqlite_filename)

	if err != nil {
		log.Fatalf("SERVE could not get bucket %s", "serve")
	}

	return nil
}
