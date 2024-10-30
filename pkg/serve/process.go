package serve

import (
	"log"

	env "search.eight/internal/env"
	"search.eight/internal/queueing"
	"search.eight/pkg/procs"
)

func Serve(ch_req chan *ServeRequest) {

	// Set up the worker.
	b, err := env.Env.GetBucket("serve")
	if err != nil {
		log.Println("cannot get bucket")
		log.Fatal(err)
	}
	s3_c := procs.NewKVS3(b)

	f, err := env.Env.GetBucket("fetch")
	if err != nil {
		log.Println("cannot get bucket")
		log.Fatal(err)
	}
	s3_f := procs.NewKVS3(f)

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[ServeRequest, ServeWorker](
		work_c, ServeRequest{},
		&ServeRequestWorker{
			ServeStorage: s3_c,
			FetchStorage: s3_f,
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
