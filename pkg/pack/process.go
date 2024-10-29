package pack

import (
	"log"

	env "search.eight/internal/env"
	"search.eight/internal/queueing"
	"search.eight/pkg/procs"
)

func Pack(ch_req chan *PackRequest) {
	// Get the K/V stores ready
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

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[PackRequest, PackWorker](
		work_c, PackRequest{},
		&PackRequestWorker{
			ExtractStorage: client_s3_extract,
			PackStorage:    client_s3_pack,
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
		work_c.Insert(job)
	}

}
