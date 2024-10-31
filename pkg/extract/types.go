package extract

import (
	env "github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/queueing"
	"github.com/jadudm/eight/pkg/procs"
	"github.com/riverqueue/river"
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
