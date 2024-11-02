package pack

import (
	"github.com/jadudm/eight/internal/queueing"
	"github.com/jadudm/eight/internal/sqlite"
	"github.com/riverqueue/river"
)

type PackRequest struct {
	Key string `json:"key"`
}

func NewPackRequest() PackRequest {
	return PackRequest{}
}

func (PackRequest) Kind() string {
	return "pack"
}

func (er PackRequest) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: er.Kind(),
	}
}

type PackRequestJob = river.Job[PackRequest]

type PackRequestWorker struct {
	EnqueueClient *queueing.River
	ChanPackages  chan Package
	ChanFinalize  chan *sqlite.PackTable

	river.WorkerDefaults[PackRequest]
}

type PackWorker = river.Worker[PackRequest]
