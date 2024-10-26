package main

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	q "search.eight/internal/queueing"
	"search.eight/pkg/crawl"
)

/* *************************** */
// The crawler looks for CrawlRequest jobs on the crawler queue.
// It exists to pick up URLs and read them into S3.
// Then, it inserts a ParseRequest job into the parser queue, so
// the file in S3 can be processed (possibly generating more CrawlRequests).
/* *************************** */

func Delay[T any](seconds int, in <-chan T, out chan<- T) {
	for {
		v := <-in
		if seconds > 0 {
			time.Sleep(time.Duration(rand.IntN(seconds)) * time.Second)
		}
		out <- v
	}

}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func main() {
	insert_delay := 0
	process_delay := 0

	var wg sync.WaitGroup
	wg.Add(1)

	ch1 := make(chan *river.Job[crawl.CrawlRequest])
	ch2 := make(chan *river.Job[crawl.CrawlRequest])

	go crawl.Crawl(ch1)
	// Add a delay in our processing
	go Delay(process_delay, ch1, ch2)

	go func(ch <-chan *river.Job[crawl.CrawlRequest]) {
		counter := 0
		for {
			<-ch
			counter += 1

			if counter%1000 == 0 {
				PrintMemUsage()
				//runtime.GC()
			}
		}
	}(ch2)

	r := q.NewRiver()
	q.QueueingClient(r, crawl.CrawlRequest{})

	// Inserts things into the queue every second
	go func() {
		o := crawl.CrawlRequest{
			Scheme: "ftp",
			Host:   "ukc.ac.uk",
			Path:   "lab",
		}
		for {
			if insert_delay > 0 {
				time.Sleep(time.Duration(rand.IntN(insert_delay)+1) * time.Millisecond)
			}

			r.Insert(o)
		}
	}()

	// Inserts things into the queue every second
	go func() {
		o := crawl.CrawlRequest{
			Scheme: "ftp",
			Host:   "ukc.ac.uk",
			Path:   uuid.NewString(),
		}
		for {
			if insert_delay > 0 {
				time.Sleep(time.Duration(rand.IntN(insert_delay)+1) * time.Millisecond)
			}

			r.Insert(o)
		}
	}()

	// FIXME: Implement a graceful shutdown from docs.
	wg.Wait()
}
