package pack

import (
	"log"

	env "search.eight/internal/env"
	"search.eight/internal/queueing"
	"search.eight/pkg/procs"
)

func Pack(ch_req chan *PackRequest) {
	// Get the K/V stores ready
	b, _ := env.Env.GetBucket("fetch")
	fb, err := env.Env.GetBucket(b.Name)
	if err != nil {
		log.Println("cannot get fetch bucket")
		log.Fatal(err)
	}

	b, _ = env.Env.GetBucket("pack")
	eb, err := env.Env.GetBucket(b.Name)
	if err != nil {
		log.Println("cannot get pack bucket")
		log.Fatal(err)
	}

	s3_fc := procs.NewKVS3(fb)
	s3_ec := procs.NewKVS3(eb)

	// This lets us queue new jobs.
	e_c := queueing.NewRiver()
	queueing.QueueingClient(e_c, PackRequest{})

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[PackRequest, PackWorker](
		work_c, PackRequest{},
		&PackRequestWorker{
			FetchStorage:  s3_fc,
			PackStorage:   s3_ec,
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
		work_c.Insert(job)
	}

}
