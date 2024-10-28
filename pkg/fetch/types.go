package fetch

import (
	"github.com/riverqueue/river"
	env "search.eight/internal/env"
	"search.eight/internal/queueing"
	"search.eight/pkg/procs"
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
	b, _ := env.Env.GetBucket("fetch")
	return b.Name
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
	StorageClient   procs.Storage
	river.WorkerDefaults[FetchRequest]
}

type FetchWorker = river.Worker[FetchRequest]
