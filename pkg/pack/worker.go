package pack

import (
	"context"
	"log"
	"sync"

	_ "modernc.org/sqlite"

	"search.eight/internal/api"
	search_db "search.eight/internal/sqlite"
	schemas "search.eight/internal/sqlite/schemas"

	"golang.org/x/sync/semaphore"
)

type Packer struct {
	Job           *PackRequestJob
	JSON          map[string]string
	PRW           *PackRequestWorker
	SearchDb      *search_db.PackTable
	Sem           *semaphore.Weighted
	WorkerContext context.Context
}

type PackionFunction func(map[string]string)

var lock = &sync.Mutex{}

var singleton_packers = sync.Map{}

func NewPacker(ctx context.Context, prw *PackRequestWorker, job *PackRequestJob) *Packer {

	json_object, err := prw.ExtractStorage.Get(job.Args.Key)
	if err != nil {
		log.Fatal(err)
	}

	host := json_object["host"]

	// Create *most* of the packer
	new_packer := &Packer{
		Job:           job,
		JSON:          json_object,
		PRW:           prw,
		WorkerContext: ctx,
	}

	// Now, add the DB connection under a singleton lock pattern.
	// We only want one DB connection for all of the workers for this domain.
	// We only want to hold the *database connection* constant, not everything else.
	// Otherwise, we get the same worker over-and-over.
	if packer, ok := singleton_packers.Load(host); ok {
		log.Println("Returning existing packer for", host)
		new_packer.SearchDb = packer.(*Packer).SearchDb
		return new_packer
	} else {
		lock.Lock()
		defer lock.Unlock()
		// Check again, now that we have the lock.
		if _, ok := singleton_packers.Load(host); ok {
			// We lost the race. Return what exists.
			log.Println("We lost the race. Using an existing packer.")
			new_packer.SearchDb = packer.(*Packer).SearchDb
			return new_packer
		} else {
			log.Println("We won the race. Creating a new  packer.", host)
			// We won the race! Create the packer.
			sqlc, err := search_db.CreatePackTable(host)

			if err != nil {
				log.Println("Error creating pack table")
				log.Fatal(err)
			}

			new_packer.SearchDb = sqlc

			singleton_packers.Store(host, new_packer)

			return new_packer
		}
	}

	log.Fatal("Could not create a packer. Should never get here.")
	return nil
}

// func (p *Packer) SequentialWrite(in chan schemas.CreateSiteEntryParams, out chan schemas.SiteIndex) {
// 	for {
// 		params := <-in
// 		ndx, err := p.SearchDb.Queries.CreateSiteEntry(p.SearchDb.Context, params)

// 		if err != nil {
// 			log.Println("Insert into site entry table failed")
// 			log.Fatal(err)
// 		}

// 		out <- ndx
// 	}

// }

func (p *Packer) Pack() {
	stats := api.NewBaseStats("pack")
	host := p.JSON["host"]

	entry_params := schemas.CreateSiteEntryParams{
		Host: host,
		Path: p.JSON["path"],
		Text: p.JSON["content"],
	}

	// Use the worker context, so all workers share the same semaphore.
	ndx, err := p.SearchDb.Queries.CreateSiteEntry(p.SearchDb.Context, entry_params)

	if err != nil {
		log.Println("Insert into site entry table failed")
		log.Fatal(err)
	}
	log.Printf("CreateSiteEntry %s %v\n", p.JSON["key"], ndx)

	stats.Increment("document_count")
	log.Println(p.JSON["path"], p.JSON["content-type"])
}

// var ch_EP chan schemas.CreateSiteEntryParams
// var ch_SI chan schemas.SiteIndex
// var one_writer = 0

func (erw *PackRequestWorker) Work(
	ctx context.Context,
	job *PackRequestJob,
) error {
	log.Println("PACK", job.Args.Key)
	// if ch_EP == nil {
	// 	ch_EP = make(chan schemas.CreateSiteEntryParams)

	// }
	// if ch_SI == nil {
	// 	ch_SI = make(chan schemas.SiteIndex)
	// }

	// Always safe to check the stats are ready.
	api.NewBaseStats("pack")

	p := NewPacker(ctx, erw, job)

	// if one_writer == 0 {
	// 	one_writer = 1
	// 	go p.SequentialWrite(ch_EP, ch_SI)
	// }

	p.Pack()

	log.Println("PACK DONE", job.Args.Key)

	return nil
}
