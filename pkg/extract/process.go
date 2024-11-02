package extract

import (
	"log"

	"github.com/jadudm/eight/internal/queueing"
)

func Extract(ch_req chan *ExtractRequest) {
	// Get the K/V stores ready

	// This lets us queue new jobs.
	e_c := queueing.NewRiver()
	queueing.QueueingClient(e_c, ExtractRequest{})

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[ExtractRequest, ExtractWorker](
		work_c, ExtractRequest{},
		&ExtractRequestWorker{
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
