package main

import (
	"sync"

	q "search.eight/internal/queueing"
	"search.eight/pkg/crawl"
)

/* *************************** */
// The crawler looks for CrawlRequest jobs on the crawler queue.
// It exists to pick up URLs and read them into S3.
// Then, it inserts a ParseRequest job into the parser queue, so
// the file in S3 can be processed (possibly generating more CrawlRequests).
/* *************************** */

func CrawlPage(ch <-chan *crawl.CrawlRequestJob) {

}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	// ch1 := make(chan *CrawlRequestJob)
	// ch2 := make(chan *river.Job[crawl.CrawlRequest])

	r := q.NewRiver()
	q.QueueingClient(r, crawl.CrawlRequest{})

	// FIXME: Implement a graceful shutdown from docs.
	wg.Wait()
}
