package pack

import (
	"github.com/riverqueue/river"
	env "search.eight/internal/env"
	"search.eight/internal/queueing"
	"search.eight/pkg/procs"
)

type PackRequest struct {
	Key string `json:"key"`
}

func NewPackRequest() PackRequest {
	return PackRequest{}
}

func (PackRequest) Kind() string {
	b, _ := env.Env.GetBucket("pack")
	return b.Name
}

func (er PackRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: er.Kind(),
	}
}

type PackRequestJob = river.Job[PackRequest]

type PackRequestWorker struct {
	FetchStorage  procs.Storage
	PackStorage   procs.Storage
	EnqueueClient *queueing.River
	river.WorkerDefaults[PackRequest]
}

type PackWorker = river.Worker[PackRequest]
