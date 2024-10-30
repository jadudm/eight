package walk

import (
	"log"
	"time"

	expirable "github.com/go-pkgz/expirable-cache/v3"
	env "search.eight/internal/env"
	"search.eight/internal/queueing"
	"search.eight/pkg/fetch"
	"search.eight/pkg/procs"
)

var cache expirable.Cache[string, int]

func get_ttl() int64 {
	ws, err := env.Env.GetService("walk")
	if err != nil {
		log.Println("WALK no service")
	}
	minutes := ws.GetParamInt64("cache-ttl-minutes")
	seconds := ws.GetParamInt64("cache-ttl-seconds")
	return (minutes * 60) + seconds
}

func Walk(ch_req chan *WalkRequest) {
	// Get the K/V stores ready
	b, _ := env.Env.GetBucket("fetch")
	fb, err := env.Env.GetBucket(b.Name)
	if err != nil {
		log.Println("cannot get fetch bucket")
		log.Fatal(err)
	}

	s3_fc := procs.NewKVS3(fb)

	// This lets us queue new jobs.
	// We have to queue things for both extract and further crawling
	e_c := queueing.NewRiver()
	queueing.QueueingClient(e_c, fetch.FetchRequest{})
	w_c := queueing.NewRiver()
	queueing.QueueingClient(w_c, WalkRequest{})

	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[WalkRequest, WalkWorker](
		work_c, WalkRequest{},
		&WalkRequestWorker{
			FetchStorage: s3_fc,
			EnqueueFetch: e_c,
			EnqueueWalk:  w_c,
		})

	if err := work_c.Client.Start(work_c.Context); err != nil {
		log.Println("Cannot start jobs")
		log.Fatal(err)
	}

	// Set up the LRU cache
	ttl := get_ttl()
	cache = expirable.NewCache[string, int]().WithTTL(time.Second * time.Duration(ttl))

	// Sit and watch for requests via the API.
	// Insert them into the queue.
	for {
		job := <-ch_req
		work_c.InsertTx(job)
	}

}
