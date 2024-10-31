package extract

import (
	"log"

	env "github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/queueing"
	"github.com/jadudm/eight/pkg/procs"
)

func Extract(ch_req chan *ExtractRequest) {
	// Get the K/V stores ready
	b, _ := env.Env.GetBucket("fetch")
	fb, err := env.Env.GetBucket(b.Name)
	if err != nil {
		log.Println("cannot get fetch bucket")
		log.Fatal(err)
	}

	b, _ = env.Env.GetBucket("extract")
	eb, err := env.Env.GetBucket(b.Name)
	if err != nil {
		log.Println("cannot get extract bucket")
		log.Fatal(err)
	}

	s3_fc := procs.NewKVS3(fb)
	s3_ec := procs.NewKVS3(eb)

	// This lets us queue new jobs.
	e_c := queueing.NewRiver()
	queueing.QueueingClient(e_c, ExtractRequest{})

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[ExtractRequest, ExtractWorker](
		work_c, ExtractRequest{},
		&ExtractRequestWorker{
			FetchStorage:   s3_fc,
			ExtractStorage: s3_ec,
			EnqueueClient:  e_c,
		})

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
