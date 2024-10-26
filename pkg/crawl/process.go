package crawl

import (
	"log"

	"github.com/riverqueue/river"
	"search.eight/internal/queueing"
)

func Crawl(out chan *river.Job[CrawlRequest]) {
	r := queueing.NewRiver()
	r = queueing.WorkingClient[CrawlRequest, river.Worker[CrawlRequest]](
		r, CrawlRequest{},
		&CrawlRequestWorker{Out: out})

	if err := r.Client.Start(r.Context); err != nil {
		log.Println("Cannot start jobs")
		log.Fatal(err)
	}
}
