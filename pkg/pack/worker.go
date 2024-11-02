package pack

import (
	"context"
	"log"
	"sync"

	_ "modernc.org/sqlite"

	kv "github.com/jadudm/eight/internal/kv"
	schemas "github.com/jadudm/eight/internal/sqlite/schemas"
)

type Package struct {
	JSON  map[string]string
	Entry schemas.CreateSiteEntryParams
}

var write_channels sync.Map

func (prw *PackRequestWorker) Work(
	ctx context.Context,
	job *PackRequestJob,
) error {
	log.Println("PACK", job.Args.Key)

	extract_storage := kv.NewKV("extract")

	obj, err := extract_storage.Get(job.Args.Key)
	if err != nil {
		log.Fatal("PACK cannot get obj from S3", job.Args.Key)
	}
	JSON := obj.GetJson()

	// Spawn a writer for each new host we see
	ch, existed := write_channels.LoadOrStore(JSON["host"], make(chan Package))
	if !existed {
		log.Println("CREATING A WRITER FOR THE HOST", JSON["host"])
		go PackWriter(ch.(chan Package), prw.ChanFinalize)
	}

	entry_params := schemas.CreateSiteEntryParams{
		Host: JSON["host"],
		Path: JSON["path"],
		Text: JSON["content"],
	}

	ch.(chan Package) <- Package{JSON, entry_params}
	return nil
}
