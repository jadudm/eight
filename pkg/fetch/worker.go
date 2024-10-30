package fetch

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"search.eight/pkg/extract"
)

func job_to_string(job *FetchRequestJob) string {
	return fmt.Sprintf("%s/%s", job.Args.Host, job.Args.Path)
}

func job_to_s3_key(job *FetchRequestJob) string {
	sha1 := sha1.Sum([]byte(job.Args.Host + job.Args.Path))
	return fmt.Sprintf("%s/%x.json", job.Args.Host, sha1)
}

func fetch_page_content(job *FetchRequestJob) map[string]string {
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

	response := map[string]string{
		"raw":            base64.URLEncoding.EncodeToString(content),
		"sha1":           fmt.Sprintf("%x", sha1.Sum(content)),
		"content-length": fmt.Sprintf("%d", len(content)),
		"host":           job.Args.Host,
		"path":           job.Args.Path,
	}

	// Copy in all of the response headers.
	for k := range res.Header {
		response[strings.ToLower(k)] = res.Header.Get(k)
	}

	return response
}

// The worker just grabs things off the queue and
// spits them out the channel. The Crawl proc then
// does the work of processing it.
func (crw *FetchRequestWorker) Work(
	ctx context.Context,
	job *FetchRequestJob,
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
		// Return with an OK status, because we don't
		// want to re-process this content yet.
		return nil
	}

	// If it is not cached, we have work to do.
	// path, err := store_to_s3(crw.Bucket, job.Args.Host, job.Args.Path)
	page_json := fetch_page_content(job)
	page_json["key"] = job_to_s3_key(job)
	err := crw.StorageClient.Store(job_to_s3_key(job), page_json)

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

	crw.EnqueueClient.InsertTx(extract.ExtractRequest{
		Key: job_to_s3_key(job),
	})

	return nil
}
