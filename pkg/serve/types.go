package serve

import (
	"github.com/riverqueue/river"
	env "search.eight/internal/env"
	"search.eight/pkg/procs"
)

type ServeRequest struct {
	Key        string `json:"key"`
	SqliteFile string `json:"sqlite_file"`
}

func NewServeRequest() ServeRequest {
	cr := ServeRequest{}
	return cr
}

func (ServeRequest) Kind() string {
	b, _ := env.Env.GetBucket("serve")
	return b.Name
}

func (cr ServeRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: cr.Kind(),
	}
}

type ServeRequestJob = river.Job[ServeRequest]

type ServeRequestWorker struct {
	ServeStorage procs.Storage
	FetchStorage procs.Storage
	river.WorkerDefaults[ServeRequest]
}

type ServeWorker = river.Worker[ServeRequest]
