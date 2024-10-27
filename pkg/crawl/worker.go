package crawl

import (
	"context"
	"fmt"
	"log"

	"github.com/riverqueue/river"
	"search.eight/internal/env"
	"search.eight/internal/queueing"
	"search.eight/pkg/cleaner"
)

type CrawlRequestWorker struct {
	CacheKeyChannel chan string
	CacheValChannel chan string
	CacheInsChannel chan map[string]string
	CleanHtmlClient *queueing.River
	Bucket          *env.Bucket

	river.WorkerDefaults[CrawlRequest]
}

// // https://github.com/philippgille/gokv?tab=readme-ov-file#usage

type CrawlWorker = river.Worker[CrawlRequest]

func job_to_string(job *CrawlRequestJob) string {
	return fmt.Sprintf("%s/%s", job.Args.Host, job.Args.Path)
}

func store_to_s3(bucket *env.Bucket, host string, path string) (string, error) {
	log.Println(bucket)
	log.Println("STORE TO S3", bucket.Credentials, host, path)
	return host + "/" + path, nil
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
	path, err := store_to_s3(crw.Bucket, job.Args.Host, job.Args.Path)
	// We get an error if we can't write to S3
	if err != nil {
		// FIXME: think about error handling from workers
		// This is just passing it back.
		log.Println(err)
		return err
	}

	// Update the cache
	crw.CacheInsChannel <- map[string]string{
		job_to_string(job): path,
	}

	// Enqueue next jobs
	crw.CleanHtmlClient.Insert(cleaner.CleanHtmlRequest{
		Bucket: "test",
		Path:   "a/b/c",
	})

	return nil
}
