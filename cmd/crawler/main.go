package main

import (
	rq "search.eight/internal/queueing"
)

type Thing struct {
	A string `json:"a"`
	B int    `json:"b"`
}

func (Thing) Kind() string { return "thingie" }

// The crawler looks for jobs on the crawler queue.
// It exists to pick up URLs and read them into S3.
// Then, it queues that page for parsing.
func main() {
	r := rq.NewRiver(rq.QueueName.Crawl)
	r.Start()
	r.Insert(rq.CrawlRequest{
		Scheme: "https",
		Host:   "jadud.com",
		Path:   "/",
	})
}
