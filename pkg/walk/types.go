package walk

import (
	"github.com/jadudm/eight/internal/queueing"
	"github.com/riverqueue/river"
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
	EnqueueFetch *queueing.River
	EnqueueWalk  *queueing.River

	river.WorkerDefaults[WalkRequest]
}

type WalkWorker = river.Worker[WalkRequest]
