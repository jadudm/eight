package main

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/google/uuid"
	"search.eight/pkg/crawl"
)

func CrawlPage(ch <-chan *crawl.CrawlRequestJob) {

}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	ch := make(chan *crawl.CrawlRequest)

	go crawl.Crawl(ch)

	for {
		time.Sleep(time.Duration(rand.IntN(10)) * time.Second)
		t := time.Now()

		ch <- &crawl.CrawlRequest{
			Host: uuid.NewString(),
			Path: t.String(),
		}
	}

	//wg.Wait()
}
