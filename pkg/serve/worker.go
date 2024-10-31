package serve

import (
	"context"
	"log"
	"os"

	env "search.eight/internal/env"
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

	s, _ := env.Env.GetService("serve")
	databases_file_path := s.GetParamString("database_files_path")

	sqlite_filename := sqlite.SqliteFilename(JSON["host"])
	// Writes to the local filesystem.
	path := databases_file_path + "/" + sqlite_filename
	log.Println(path, "<-", sqlite_filename, os.Getenv("PWD"))
	srw.ServeStorage.GetObject(path, sqlite_filename)

	if err != nil {
		log.Fatalf("SERVE could not get bucket %s", "serve")
	}

	return nil
}
