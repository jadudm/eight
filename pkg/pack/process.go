package pack

import (
	"context"
	"log"
	"time"

	"search.eight/internal/api"
	env "search.eight/internal/env"
	"search.eight/internal/queueing"
	sqlite "search.eight/internal/sqlite"
	"search.eight/internal/util"
	"search.eight/pkg/procs"
)

// The PackWriter provides concurrency protection for the
// SQLite databases. We can only write one package at a time.
// This is because we have 10 workers.
func PackWriter(chp <-chan Package, chf chan<- *sqlite.PackTable) {
	databases := make(map[string]*sqlite.PackTable)
	contexts := make(map[string]context.Context)
	stats := api.NewBaseStats("pack")

	for {
		pkg := <-chp
		host := pkg.JSON["host"]
		// log.Println("PACKING", host, pkg.JSON["key"])
		// Only create the connection once.
		if _, ok := databases[host]; !ok {
			table, err := sqlite.CreatePackTable(sqlite.SqliteFilename(host), pkg.JSON)
			if err != nil {
				log.Println("Could not create pack table for", host)
				log.Fatal(err)
			}
			databases[host] = table
			contexts[host] = context.Background()
		}

		_, err := databases[host].Queries.CreateSiteEntry(contexts[host], pkg.Entry)

		if err != nil {
			log.Println("Insert into site entry table failed")
			log.Fatal(err)
		}

		//log.Printf("CreateSiteEntry %s %v\n", pkg.JSON["key"], si.Path)
		stats.Increment("document_count")
		chf <- databases[host]
	}
}

// We get pings on domains as they go through
// When the timer fires, we queue that domain to the finalize queue.
func FinalizeTimer(in <-chan *sqlite.PackTable) {
	//FIXME: This should be a config parameter
	TIMEOUT_DURATION := time.Duration(10 * time.Second)

	clocks := make(map[string]time.Time)
	tables := make(map[string]*sqlite.PackTable)

	// https://dev.to/milktea02/misunderstanding-go-timers-and-channels-1jal
	log.Println("FINALIZE starting timer...")
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
			for sqlite_filename, clock := range clocks {
				if time.Since(clock) > TIMEOUT_DURATION {
					//prw.EnqueueClient()
					// FIXME: Just send it to S3 for now.
					// This is still a bit of an MVP.

					b, err := env.Env.GetBucket("serve")
					if err != nil {
						log.Println("cannot get serve bucket")
						log.Fatal(err)
					}
					kv := procs.NewKVS3(b)
					log.Println("FINALIZE streaming", sqlite_filename)

					tables[sqlite_filename].PrepForNetwork()

					err = kv.StreamObject(sqlite_filename, sqlite_filename)
					if err != nil {
						log.Fatal(err)
					}

					//tables[sqlite_filename].DB.Close()
					log.Println("PACK", tables[sqlite_filename].JSON)
					// Enqueue serve
					// This generic queue lets us queue new jobs
					// when we don't have another handle to grab.
					e_c := queueing.NewRiver()
					queueing.QueueingClient(e_c, util.GenericRequest{})
					e_c.InsertTx(util.GenericRequest{
						Key:       tables[sqlite_filename].JSON["key"],
						QueueName: "serve"})

					delete(clocks, sqlite_filename)
					delete(tables, sqlite_filename)

				}
			}
			timeout.Reset(TIMEOUT_DURATION)
		}
	}
}

func Pack(ch_req chan *PackRequest) {
	// Spin up the helper processes
	ch_packages := make(chan Package)
	ch_finalize := make(chan *sqlite.PackTable)

	// FIXME: we need a finalize client in here
	go FinalizeTimer(ch_finalize)
	//go PackWriter(ch_packages, ch_finalize)

	b, _ := env.Env.GetBucket("extract")
	extract_b, err := env.Env.GetBucket(b.Name)
	if err != nil {
		log.Println("cannot get fetch bucket")
		log.Fatal(err)
	}

	b, _ = env.Env.GetBucket("pack")
	pack_b, err := env.Env.GetBucket(b.Name)
	if err != nil {
		log.Println("cannot get pack bucket")
		log.Fatal(err)
	}

	client_s3_extract := procs.NewKVS3(extract_b)
	client_s3_pack := procs.NewKVS3(pack_b)

	// This lets us queue new jobs.
	e_c := queueing.NewRiver()
	queueing.QueueingClient(e_c, PackRequest{})

	prw := &PackRequestWorker{
		ExtractStorage: client_s3_extract,
		PackStorage:    client_s3_pack,
		EnqueueClient:  e_c,
		ChanPackages:   ch_packages,
		ChanFinalize:   ch_finalize,
	}
	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[PackRequest, PackWorker](
		work_c, PackRequest{}, prw)

	if err := work_c.Client.Start(work_c.Context); err != nil {
		log.Println("Cannot start jobs")
		log.Fatal(err)
	}

	// Sit and watch for requests via the API.
	// Insert them into the queue.
	for {
		job := <-ch_req
		work_c.InsertTx(job)
	}

}
