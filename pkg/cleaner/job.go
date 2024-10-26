package cleaner

import "github.com/riverqueue/river"

type CleanHtmlRequest struct {
	Bucket string `json:"bucket"`
	Path   string `json:"path"`
}

func (CleanHtmlRequest) Kind() string { return "clean_html" }

func (cr CleanHtmlRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: cr.Kind(),
	}
}

type CleanHtmlRequestJob = river.Job[CleanHtmlRequest]
