package crawl

import "github.com/riverqueue/river"

type CrawlRequest struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

func (CrawlRequest) Kind() string { return "crawl" }

func (cr CrawlRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: cr.Kind(),
	}
}
