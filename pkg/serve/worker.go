package serve

import (
	"context"
	"log"
	"os"

	env "github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/sqlite"
	kv "github.com/jadudm/eight/pkg/kv"
)

// The worker just grabs things off the queue and
// spits them out the channel. The Crawl proc then
// does the work of processing it.
func (srw *ServeRequestWorker) Work(
	ctx context.Context,
	job *ServeRequestJob,
) error {
	fetch_storage := kv.NewKV("fetch")
	serve_storage := kv.NewKV("serve")

	// Our requests will be for packed DBs.
	// So, those should be fetched into the local system, and
	// the API should know that we can now serve the domain.
	obj, err := fetch_storage.Get(job.Args.Key)
	if err != nil {
		log.Fatal(err)
	}
	JSON := obj.GetJson()

	s, _ := env.Env.GetUserService("serve")
	databases_file_path := s.GetParamString("database_files_path")

	sqlite_filename := sqlite.SqliteFilename(JSON["host"])
	// Writes to the local filesystem.
	destination := databases_file_path + "/" + sqlite_filename
	log.Println(destination, "<-", sqlite_filename, os.Getenv("PWD"))

	err = serve_storage.GetFile(sqlite_filename, destination)

	//srw.ServeStorage.GetObject(path, sqlite_filename)

	if err != nil {
		log.Fatalf("SERVE could not get bucket %s", "serve")
	}

	return nil
}
