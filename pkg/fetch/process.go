package fetch

import (
	"log"

	"github.com/jadudm/eight/internal/queueing"
	"github.com/jadudm/eight/pkg/extract"
)

func Fetch(ch_req chan *FetchRequest) {

	ch_key := make(chan string)
	ch_val := make(chan string)
	ch_ins := make(chan map[string]string)

	// This lets us queue new jobs.
	clean_c := queueing.NewRiver()
	queueing.QueueingClient(clean_c, extract.ExtractRequest{})

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[FetchRequest, FetchWorker](
		work_c, FetchRequest{},
		&FetchRequestWorker{
			CacheKeyChannel: ch_key,
			CacheValChannel: ch_val,
			CacheInsChannel: ch_ins,
			EnqueueClient:   clean_c,
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
