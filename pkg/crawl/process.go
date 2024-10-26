package crawl

import (
	"log"

	"search.eight/internal/queueing"
)

func Crawl(out chan *CrawlRequestJob) {
	r := queueing.NewRiver()
	r = queueing.WorkingClient[CrawlRequest, CrawlWorker](
		r, CrawlRequest{},
		&CrawlRequestWorker{Out: out})

	if err := r.Client.Start(r.Context); err != nil {
		log.Println("Cannot start jobs")
		log.Fatal(err)
	}
}
