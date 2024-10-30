package walk

import (
	"github.com/riverqueue/river"
	"search.eight/internal/queueing"
	"search.eight/pkg/procs"
)

type WalkRequest struct {
	Key string `json:"key"`
}

func NewWalkRequest() WalkRequest {
	return WalkRequest{}
}

func (WalkRequest) Kind() string {
	return "walk"
}

func (er WalkRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: er.Kind(),
	}
}

type WalkRequestJob = river.Job[WalkRequest]

type WalkRequestWorker struct {
	FetchStorage procs.Storage
	EnqueueFetch *queueing.River
	EnqueueWalk  *queueing.River

	river.WorkerDefaults[WalkRequest]
}

type WalkWorker = river.Worker[WalkRequest]
