package fetch

import (
	env "github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/queueing"
	"github.com/jadudm/eight/pkg/procs"
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
