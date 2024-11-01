package extract

import (
	"log"

	env "github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/queueing"
	"github.com/jadudm/eight/pkg/procs"
)

func Extract(ch_req chan *ExtractRequest) {
	// Get the K/V stores ready
	b, _ := env.Env.GetBucket(env.WorkingObjectStore)
	s3_b := procs.NewKVS3(b)

	// This lets us queue new jobs.
	e_c := queueing.NewRiver()
	queueing.QueueingClient(e_c, ExtractRequest{})

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[ExtractRequest, ExtractWorker](
		work_c, ExtractRequest{},
		&ExtractRequestWorker{
			ObjectStorage: s3_b,
			EnqueueClient: e_c,
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
