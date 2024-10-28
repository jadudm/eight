package extract

import (
	"github.com/riverqueue/river"
	env "search.eight/internal/env"
	"search.eight/internal/queueing"
	"search.eight/pkg/procs"
)

type ExtractRequest struct {
	Key string `json:"key"`
}

func NewExtractRequest() ExtractRequest {
	return ExtractRequest{}
}

func (ExtractRequest) Kind() string {
	b, _ := env.Env.GetBucket("extract")
	return b.Name
}

func (er ExtractRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: er.Kind(),
	}
}

type ExtractRequestJob = river.Job[ExtractRequest]

type ExtractRequestWorker struct {
	FetchStorage   procs.Storage
	ExtractStorage procs.Storage
	EnqueueClient  *queueing.River
	river.WorkerDefaults[ExtractRequest]
}

type ExtractWorker = river.Worker[ExtractRequest]
