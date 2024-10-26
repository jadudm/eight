package crawl

import (
	"log"

	"search.eight/internal/queueing"
	"search.eight/pkg/cleaner"
	"search.eight/pkg/procs"
)

func Crawl(ch_req chan *CrawlRequest) {

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
	work_c := queueing.NewRiver()
	work_c = queueing.WorkingClient[CrawlRequest, CrawlWorker](
		work_c, CrawlRequest{},
		&CrawlRequestWorker{
			CacheKeyChannel: ch_key,
			CacheValChannel: ch_val,
			CacheInsChannel: ch_ins,
			CleanHtmlClient: clean_c,
		})

	if err := work_c.Client.Start(work_c.Context); err != nil {
		log.Println("Cannot start jobs")
		log.Fatal(err)
	}

	for {
		job := <-ch_req
		work_c.Insert(job)
	}

}
