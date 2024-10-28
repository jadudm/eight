package fetch

import (
	"log"

	env "search.eight/internal/env"
	"search.eight/internal/queueing"
	"search.eight/pkg/cleaner"
	"search.eight/pkg/procs"
)

func Fetch(ch_req chan *FetchRequest) {

	ch_key := make(chan string)
	ch_val := make(chan string)
	ch_ins := make(chan map[string]string)

	// Run the cache process.
	// FIXME: make it so we can configure the Crawl proc
	// with a cache. perhaps by passing in a
	// cache channel bundle Cache{chan, chan, chan}
	go procs.StringCache(ch_key, ch_val, ch_ins)

	// This lets us queue new jobs.
	clean_c := queueing.NewRiver()
	queueing.QueueingClient(clean_c, cleaner.CleanHtmlRequest{})

	// Set up the worker.
	b, err := env.Env.GetBucket(const_bucket_s3)

	if err != nil {
		log.Println("cannot get bucket")
		log.Fatal(err)
	}

	s3_c := procs.NewKVS3(b)

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[FetchRequest, FetchWorker](
		work_c, FetchRequest{},
		&FetchRequestWorker{
			CacheKeyChannel: ch_key,
			CacheValChannel: ch_val,
			CacheInsChannel: ch_ins,
			CleanHtmlClient: clean_c,
			StorageClient:   s3_c,
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
