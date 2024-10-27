package crawl

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/riverqueue/river"
	"search.eight/internal/queueing"
	"search.eight/pkg/procs"
)

type CrawlRequestWorker struct {
	CacheKeyChannel chan string
	CacheValChannel chan string
	CacheInsChannel chan map[string]string
	CleanHtmlClient *queueing.River
	StorageClient   procs.Storage

	river.WorkerDefaults[CrawlRequest]
}

type CrawlWorker = river.Worker[CrawlRequest]

func job_to_string(job *CrawlRequestJob) string {
	return fmt.Sprintf("%s/%s", job.Args.Host, job.Args.Path)
}

func job_to_s3_key(job *CrawlRequestJob) string {
	sha1 := sha1.Sum([]byte(job.Args.Path))
	return fmt.Sprintf("%s/%x\n", job.Args.Host, sha1)

}

func fetch_page_content(job *CrawlRequestJob) map[string]string {
	url := url.URL{
		Scheme: job.Args.Scheme,
		Host:   job.Args.Host,
		Path:   job.Args.Path,
	}

	res, err := http.Get(url.String())
	if err != nil {
		log.Fatal(err)
	}

	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	return map[string]string{
		"content": string(content),
	}
}

// The worker just grabs things off the queue and
// spits them out the channel. The Crawl proc then
// does the work of processing it.
func (crw *CrawlRequestWorker) Work(
	ctx context.Context,
	job *CrawlRequestJob,
) error {

	// Check the cache.
	// Using channels because we don't know how/where the cache is
	// implemented, and we just want to send/receive results.
	crw.CacheKeyChannel <- job_to_string(job)
	path_s3 := <-crw.CacheValChannel

	// If it is already cached, we have nothing to do.
	// We'll queue the cleaner. All we know at this point
	// is that it was pulled down, and we have it in S3.
	if path_s3 != "" {
		// Go back to the top of the loop and wait
		// for something else on the channel.
		return nil
	}

	// If it is not cached, we have work to do.
	// path, err := store_to_s3(crw.Bucket, job.Args.Host, job.Args.Path)
	page_bytes := fetch_page_content(job)

	err := crw.StorageClient.Store(job_to_s3_key(job), page_bytes)

	// We get an error if we can't write to S3
	if err != nil {
		log.Println("could not store k/v")
		log.Println(err)
		return err
	}

	// Update the cache
	crw.CacheInsChannel <- map[string]string{
		job_to_string(job): path_s3,
	}

	// // Enqueue next jobs
	// crw.CleanHtmlClient.Insert(cleaner.CleanHtmlRequest{
	// 	Bucket: "test",
	// 	Path:   "a/b/c",
	// })

	return nil
}
