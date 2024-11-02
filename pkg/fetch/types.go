package fetch

import (
	"github.com/jadudm/eight/internal/queueing"
	"github.com/riverqueue/river"
)

type FetchRequest struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

func NewFetchRequest() FetchRequest {
	cr := FetchRequest{}
	cr.Scheme = "https"
	return cr
}

func (FetchRequest) Kind() string {
	return "fetch"
}

func (cr FetchRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: cr.Kind(),
	}
}

type FetchRequestJob = river.Job[FetchRequest]

type FetchRequestWorker struct {
	CacheKeyChannel chan string
	CacheValChannel chan string
	CacheInsChannel chan map[string]string
	EnqueueClient   *queueing.River
	river.WorkerDefaults[FetchRequest]
}

type FetchWorker = river.Worker[FetchRequest]
