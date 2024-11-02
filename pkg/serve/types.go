package serve

import (
	"github.com/riverqueue/river"
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
	return "serve"
}

func (cr ServeRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: cr.Kind(),
	}
}

type ServeRequestJob = river.Job[ServeRequest]

type ServeRequestWorker struct {
	river.WorkerDefaults[ServeRequest]
}

type ServeWorker = river.Worker[ServeRequest]
