package serve

import (
	"log"

	"github.com/jadudm/eight/internal/queueing"
)

func Serve(ch_req chan *ServeRequest) {

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[ServeRequest, ServeWorker](
		work_c, ServeRequest{},
		&ServeRequestWorker{})

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
