package crawl

import (
	"github.com/riverqueue/river"
	"search.eight/internal/queueing"
	"search.eight/pkg/procs"
)

var const_bucket_s3 = "crawl"

type CrawlRequest struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

func NewCrawlRequest() CrawlRequest {
	cr := CrawlRequest{}
	cr.Scheme = "https"
	return cr
}

func (CrawlRequest) Kind() string { return "crawl" }

func (cr CrawlRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: cr.Kind(),
	}
}

type CrawlRequestJob = river.Job[CrawlRequest]

type CrawlRequestWorker struct {
	CacheKeyChannel chan string
	CacheValChannel chan string
	CacheInsChannel chan map[string]string
	CleanHtmlClient *queueing.River
	StorageClient   procs.Storage

	river.WorkerDefaults[CrawlRequest]
}

type CrawlWorker = river.Worker[CrawlRequest]
