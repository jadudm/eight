package extract

import (
	kv "github.com/jadudm/eight/internal/kv"
	"github.com/jadudm/eight/internal/queueing"
	"github.com/riverqueue/river"
)

type ExtractRequest struct {
	Key string `json:"key"`
}

func NewExtractRequest() ExtractRequest {
	return ExtractRequest{}
}

func (ExtractRequest) Kind() string {
	return "extract"
}

func (er ExtractRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: er.Kind(),
	}
}

type ExtractRequestJob = river.Job[ExtractRequest]

type ExtractRequestWorker struct {
	ObjectStorage kv.Storage
	EnqueueClient *queueing.River
	river.WorkerDefaults[ExtractRequest]
}

type ExtractWorker = river.Worker[ExtractRequest]
