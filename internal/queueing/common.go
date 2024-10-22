package queueing

type CrawlRequest struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

func (CrawlRequest) Kind() string { return "crawl_request" }
